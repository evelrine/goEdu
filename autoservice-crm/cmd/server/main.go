package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/phpdave11/gofpdf"
	"golang.org/x/crypto/bcrypt"
)

// App хранит общие зависимости приложения: подключение к БД, шаблоны и сессии.
type App struct {
	db        *pgxpool.Pool
	templates *template.Template
	store     *sessions.CookieStore
}

// TemplateData — единая структура данных, которую передаем в HTML-шаблоны.
type TemplateData struct {
	Title   string
	User    string
	Error   string
	Success string
	Data    any
}

type Client struct {
	ID       int
	FullName string
	Phone    string
	Email    string
	Notes    string
}

type Car struct {
	ID           int
	ClientID     int
	ClientName   string
	Brand        string
	Model        string
	Year         int
	VIN          string
	LicensePlate string
	Mileage      int
}

type Order struct {
	ID                 int
	ClientID           int
	CarID              int
	ClientName         string
	CarTitle           string
	Status             string
	ProblemDescription string
	ManagerComment     string
	TotalPrice         float64
	CreatedAt          time.Time
}

type OrderItem struct {
	ID       int
	Type     string
	Title    string
	Quantity float64
	Price    float64
	Total    float64
}

func main() {
	// DATABASE_URL приходит из Docker Compose; fallback нужен для локального запуска без Docker.
	databaseURL := env("DATABASE_URL", "postgres://autoservice:autoservice_password@localhost:5432/autoservice?sslmode=disable")
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// При запуске через Docker база может стартовать чуть позже приложения.
	for i := 0; i < 20; i++ {
		if err := pool.Ping(context.Background()); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	// Регистрируем функции для шаблонов и загружаем все HTML-файлы интерфейса.
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"statusLabel": statusLabel,
	}).ParseGlob("web/templates/*.html"))
	secret := env("APP_SECRET", "local-secret-change-me")
	// Сессия хранит id и имя пользователя после входа. Secure=true подходит для HTTPS через Caddy.
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{Path: "/", MaxAge: 86400 * 7, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}

	app := &App{db: pool, templates: tmpl, store: store}

	// Маршруты приложения. Все рабочие разделы закрыты middleware requireAuth.
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	mux.HandleFunc("/", app.requireAuth(app.dashboard))
	mux.HandleFunc("/login", app.login)
	mux.HandleFunc("/register", app.register)
	mux.HandleFunc("/logout", app.logout)
	mux.HandleFunc("/clients", app.requireAuth(app.clients))
	mux.HandleFunc("/cars", app.requireAuth(app.cars))
	mux.HandleFunc("/orders", app.requireAuth(app.orders))
	mux.HandleFunc("/orders/", app.requireAuth(app.orderAction))

	port := env("APP_PORT", "8080")
	log.Println("AutoService CRM started on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// statusLabel переводит технический статус из БД в понятный текст для интерфейса и PDF.
func statusLabel(status string) string {
	switch status {
	case "new":
		return "Новый"
	case "in_progress":
		return "В работе"
	case "waiting_parts":
		return "Ожидает запчасти"
	case "done":
		return "Завершен"
	case "cancelled":
		return "Отменен"
	default:
		return status
	}
}

// isValidOrderStatus защищает от записи произвольного статуса в базу.
func isValidOrderStatus(status string) bool {
	switch status {
	case "new", "in_progress", "waiting_parts", "done", "cancelled":
		return true
	default:
		return false
	}
}

func env(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// render добавляет текущего пользователя и выводит нужный HTML-шаблон.
func (a *App) render(w http.ResponseWriter, r *http.Request, name string, td TemplateData) {
	td.User = a.currentUser(r)
	if err := a.templates.ExecuteTemplate(w, name, td); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) session(r *http.Request) *sessions.Session {
	s, _ := a.store.Get(r, "autoservice_session")
	return s
}

func (a *App) currentUser(r *http.Request) string {
	s := a.session(r)
	v, _ := s.Values["user_name"].(string)
	return v
}

// requireAuth не пускает в CRM без авторизации.
func (a *App) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := a.session(r)
		if s.Values["user_id"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// register обрабатывает страницу регистрации и сохраняет пароль только в виде bcrypt-хеша.
func (a *App) register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.render(w, r, "register.html", TemplateData{Title: "Регистрация"})
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	if name == "" || email == "" || len(password) < 6 {
		a.render(w, r, "register.html", TemplateData{Title: "Регистрация", Error: "Заполните поля. Пароль минимум 6 символов."})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := a.db.Exec(r.Context(), "INSERT INTO users (name,email,password_hash) VALUES ($1,$2,$3)", name, email, string(hash))
	if err != nil {
		a.render(w, r, "register.html", TemplateData{Title: "Регистрация", Error: "Пользователь с таким email уже существует."})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// login проверяет email/пароль и создает пользовательскую сессию.
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.render(w, r, "login.html", TemplateData{Title: "Вход"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	var id int
	var name, hash string
	err := a.db.QueryRow(r.Context(), "SELECT id,name,password_hash FROM users WHERE email=$1", email).Scan(&id, &name, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		a.render(w, r, "login.html", TemplateData{Title: "Вход", Error: "Неверный email или пароль."})
		return
	}
	s := a.session(r)
	s.Values["user_id"] = id
	s.Values["user_name"] = name
	_ = s.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	s := a.session(r)
	s.Options.MaxAge = -1
	_ = s.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// dashboard собирает краткую статистику для главной страницы.
func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	stats := map[string]any{}
	var clientsCount, carsCount, ordersCount int
	var revenue float64
	_ = a.db.QueryRow(r.Context(), "SELECT count(*) FROM clients").Scan(&clientsCount)
	_ = a.db.QueryRow(r.Context(), "SELECT count(*) FROM cars").Scan(&carsCount)
	_ = a.db.QueryRow(r.Context(), "SELECT count(*) FROM orders").Scan(&ordersCount)
	_ = a.db.QueryRow(r.Context(), "SELECT coalesce(sum(total_price),0) FROM orders WHERE status='done'").Scan(&revenue)
	stats["clients"] = clientsCount
	stats["cars"] = carsCount
	stats["orders"] = ordersCount
	stats["revenue"] = revenue
	rows, _ := a.db.Query(r.Context(), `SELECT o.id, c.full_name, cars.brand || ' ' || cars.model, o.status, o.total_price, o.created_at FROM orders o JOIN clients c ON c.id=o.client_id JOIN cars ON cars.id=o.car_id ORDER BY o.created_at DESC LIMIT 5`)
	defer rows.Close()
	latest := []Order{}
	for rows.Next() {
		var o Order
		_ = rows.Scan(&o.ID, &o.ClientName, &o.CarTitle, &o.Status, &o.TotalPrice, &o.CreatedAt)
		latest = append(latest, o)
	}
	stats["latest"] = latest
	a.render(w, r, "dashboard.html", TemplateData{Title: "Панель", Data: stats})
}

// clients показывает список клиентов и обрабатывает добавление нового клиента.
func (a *App) clients(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		_, err := a.db.Exec(r.Context(), "INSERT INTO clients(full_name,phone,email,notes) VALUES($1,$2,$3,$4)", r.FormValue("full_name"), r.FormValue("phone"), r.FormValue("email"), r.FormValue("notes"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, "/clients", http.StatusSeeOther)
		return
	}
	rows, _ := a.db.Query(r.Context(), "SELECT id,full_name,phone,email,notes FROM clients ORDER BY id DESC")
	defer rows.Close()
	items := []Client{}
	for rows.Next() {
		var c Client
		_ = rows.Scan(&c.ID, &c.FullName, &c.Phone, &c.Email, &c.Notes)
		items = append(items, c)
	}
	a.render(w, r, "clients.html", TemplateData{Title: "Клиенты", Data: items})
}

// cars показывает автомобили и позволяет привязать новый автомобиль к клиенту.
func (a *App) cars(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		clientID, _ := strconv.Atoi(r.FormValue("client_id"))
		year, _ := strconv.Atoi(r.FormValue("year"))
		mileage, _ := strconv.Atoi(r.FormValue("mileage"))
		_, err := a.db.Exec(r.Context(), "INSERT INTO cars(client_id,brand,model,year,vin,license_plate,mileage) VALUES($1,$2,$3,$4,$5,$6,$7)", clientID, r.FormValue("brand"), r.FormValue("model"), year, r.FormValue("vin"), r.FormValue("license_plate"), mileage)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, "/cars", http.StatusSeeOther)
		return
	}
	clients := a.getClients(r.Context())
	rows, _ := a.db.Query(r.Context(), `SELECT cars.id, client_id, clients.full_name, brand, model, year, vin, license_plate, mileage FROM cars JOIN clients ON clients.id=cars.client_id ORDER BY cars.id DESC`)
	defer rows.Close()
	cars := []Car{}
	for rows.Next() {
		var c Car
		_ = rows.Scan(&c.ID, &c.ClientID, &c.ClientName, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.LicensePlate, &c.Mileage)
		cars = append(cars, c)
	}
	a.render(w, r, "cars.html", TemplateData{Title: "Автомобили", Data: map[string]any{"cars": cars, "clients": clients}})
}

// orders отвечает за список заказ-нарядов и создание нового заказ-наряда.
func (a *App) orders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		clientID, _ := strconv.Atoi(r.FormValue("client_id"))
		carID, _ := strconv.Atoi(r.FormValue("car_id"))

		// Заказ-наряд можно создать только для автомобиля выбранного клиента.
		// Проверка на backend обязательна, потому что HTML/JS можно обойти вручную.
		if !a.carBelongsToClient(r.Context(), carID, clientID) {
			a.render(w, r, "orders.html", TemplateData{
				Title: "Заказ-наряды",
				Error: "Выбранный автомобиль не принадлежит указанному клиенту.",
				Data:  a.ordersPageData(r.Context()),
			})
			return
		}

		serviceTitle := r.FormValue("service_title")
		servicePrice, _ := strconv.ParseFloat(r.FormValue("service_price"), 64)
		partTitle := r.FormValue("part_title")
		partPrice, _ := strconv.ParseFloat(r.FormValue("part_price"), 64)
		total := servicePrice + partPrice
		var orderID int
		err := a.db.QueryRow(r.Context(), "INSERT INTO orders(client_id,car_id,status,problem_description,manager_comment,total_price) VALUES($1,$2,$3,$4,$5,$6) RETURNING id", clientID, carID, r.FormValue("status"), r.FormValue("problem_description"), r.FormValue("manager_comment"), total).Scan(&orderID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if serviceTitle != "" {
			_, _ = a.db.Exec(r.Context(), "INSERT INTO order_items(order_id,item_type,title,quantity,price) VALUES($1,'service',$2,1,$3)", orderID, serviceTitle, servicePrice)
		}
		if partTitle != "" {
			_, _ = a.db.Exec(r.Context(), "INSERT INTO order_items(order_id,item_type,title,quantity,price) VALUES($1,'part',$2,1,$3)", orderID, partTitle, partPrice)
		}
		http.Redirect(w, r, fmt.Sprintf("/orders/%d", orderID), http.StatusSeeOther)
		return
	}
	rows, _ := a.db.Query(r.Context(), `SELECT o.id, c.full_name, cars.brand || ' ' || cars.model || ' (' || coalesce(cars.license_plate,'') || ')', o.status, o.total_price, o.created_at FROM orders o JOIN clients c ON c.id=o.client_id JOIN cars ON cars.id=o.car_id ORDER BY o.id DESC`)
	defer rows.Close()
	orders := []Order{}
	for rows.Next() {
		var o Order
		_ = rows.Scan(&o.ID, &o.ClientName, &o.CarTitle, &o.Status, &o.TotalPrice, &o.CreatedAt)
		orders = append(orders, o)
	}
	data := a.ordersPageData(r.Context())
	data["orders"] = orders
	a.render(w, r, "orders.html", TemplateData{Title: "Заказ-наряды", Data: data})
}

// ordersPageData собирает справочники для страницы создания заказ-наряда.
func (a *App) ordersPageData(ctx context.Context) map[string]any {
	return map[string]any{
		"orders":  []Order{},
		"clients": a.getClients(ctx),
		"cars":    a.getCars(ctx),
	}
}

// carBelongsToClient проверяет связь автомобиля и клиента перед созданием заказа.
func (a *App) carBelongsToClient(ctx context.Context, carID, clientID int) bool {
	if carID <= 0 || clientID <= 0 {
		return false
	}
	var exists bool
	err := a.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM cars WHERE id=$1 AND client_id=$2)", carID, clientID).Scan(&exists)
	return err == nil && exists
}

// orderAction разбирает вложенные действия заказа: просмотр, PDF и изменение статуса.
func (a *App) orderAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	id, _ := strconv.Atoi(parts[0])
	if len(parts) > 1 {
		switch parts[1] {
		case "pdf":
			a.orderPDF(w, r, id)
			return
		case "status":
			a.orderStatusUpdate(w, r, id)
			return
		}
	}
	a.orderShow(w, r, id)
}

func (a *App) orderShow(w http.ResponseWriter, r *http.Request, id int) {
	o, items, err := a.loadOrder(r.Context(), id)
	if err != nil {
		http.Error(w, "Заказ-наряд не найден", http.StatusNotFound)
		return
	}
	a.render(w, r, "order_show.html", TemplateData{Title: fmt.Sprintf("Заказ №%d", id), Data: map[string]any{"order": o, "items": items}})
}

// orderStatusUpdate меняет статус заказ-наряда из формы на странице заказа.
func (a *App) orderStatusUpdate(w http.ResponseWriter, r *http.Request, id int) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status := r.FormValue("status")
	if !isValidOrderStatus(status) {
		http.Error(w, "invalid order status", http.StatusBadRequest)
		return
	}
	_, err := a.db.Exec(r.Context(), "UPDATE orders SET status=$1, completed_at=CASE WHEN $1='done' THEN now() ELSE NULL END WHERE id=$2", status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/orders/%d", id), http.StatusSeeOther)
}

// orderPDF формирует печатную PDF-версию заказ-наряда.
func (a *App) orderPDF(w http.ResponseWriter, r *http.Request, id int) {
	o, items, err := a.loadOrder(r.Context(), id)
	if err != nil {
		http.Error(w, "Заказ-наряд не найден", http.StatusNotFound)
		return
	}

	// Для кириллицы нужен TTF-шрифт. В Docker он ставится пакетом ttf-dejavu.
	fontPath, err := findPDFFont()
	if err != nil {
		log.Printf("pdf font not found: %v", err)
		http.Error(w, "Не найден TTF-шрифт для PDF. Запустите проект через Docker или укажите PDF_FONT_PATH.", http.StatusInternalServerError)
		return
	}

	// Загружаем шрифт байтами, а не передаем путь в gofpdf.
	// Так PDF стабильно формируется в Docker, потому что библиотека не меняет абсолютный путь к файлу.
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		log.Printf("pdf font read error: %v", err)
		http.Error(w, "Не удалось прочитать шрифт для PDF", http.StatusInternalServerError)
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddUTF8FontFromBytes("DejaVu", "", fontBytes)
	pdf.AddPage()

	pdf.SetFont("DejaVu", "", 18)
	pdf.Cell(180, 10, "AutoService CRM")
	pdf.Ln(12)

	pdf.SetFont("DejaVu", "", 14)
	pdf.Cell(180, 9, fmt.Sprintf("Заказ-наряд №%d от %s", o.ID, o.CreatedAt.Format("02.01.2006")))
	pdf.Ln(12)

	pdf.SetFont("DejaVu", "", 11)
	pdf.Cell(180, 7, "Клиент: "+o.ClientName)
	pdf.Ln(7)
	pdf.Cell(180, 7, "Автомобиль: "+o.CarTitle)
	pdf.Ln(7)
	pdf.Cell(180, 7, "Статус: "+statusLabel(o.Status))
	pdf.Ln(9)

	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(180, 8, "Описание проблемы")
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 10)
	pdf.MultiCell(180, 6, emptyDash(o.ProblemDescription), "", "L", false)
	pdf.Ln(4)

	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(180, 8, "Позиции")
	pdf.Ln(8)
	pdf.SetFont("DejaVu", "", 10)
	pdf.CellFormat(85, 8, "Позиция", "1", 0, "L", false, 0, "")
	pdf.CellFormat(25, 8, "Кол-во", "1", 0, "C", false, 0, "")
	pdf.CellFormat(35, 8, "Цена", "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, "Итого", "1", 1, "R", false, 0, "")

	if len(items) == 0 {
		pdf.CellFormat(180, 8, "Позиции не добавлены", "1", 1, "L", false, 0, "")
	} else {
		for _, it := range items {
			pdf.CellFormat(85, 8, it.Title, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", it.Quantity), "1", 0, "C", false, 0, "")
			pdf.CellFormat(35, 8, fmt.Sprintf("%.2f руб.", it.Price), "1", 0, "R", false, 0, "")
			pdf.CellFormat(35, 8, fmt.Sprintf("%.2f руб.", it.Total), "1", 1, "R", false, 0, "")
		}
	}

	pdf.Ln(6)
	pdf.SetFont("DejaVu", "", 13)
	pdf.CellFormat(180, 9, fmt.Sprintf("Итого: %.2f руб.", o.TotalPrice), "", 1, "R", false, 0, "")
	pdf.Ln(12)

	pdf.SetFont("DejaVu", "", 10)
	pdf.Cell(90, 8, "Подпись клиента: __________________")
	pdf.Cell(90, 8, "Подпись мастера: __________________")

	// Сначала пишем PDF в буфер. Так пользователь не получит пустой файл при ошибке.
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		log.Printf("pdf generation error for order %d: %v", id, err)
		http.Error(w, "Не удалось сформировать PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=order-%d.pdf", id))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}

// findPDFFont ищет шрифт с поддержкой кириллицы в Docker, macOS или по PDF_FONT_PATH.
func findPDFFont() (string, error) {
	if custom := os.Getenv("PDF_FONT_PATH"); custom != "" {
		if _, err := os.Stat(custom); err == nil {
			return custom, nil
		}
	}

	candidates := []string{
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/ttf-dejavu/DejaVuSans.ttf",
		"/usr/local/share/fonts/DejaVuSans.ttf",
		"/Library/Fonts/DejaVuSans.ttf",
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("pdf font not found")
}

// emptyDash нужен, чтобы в PDF не было пустых блоков.
func emptyDash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	return s
}

// loadOrder загружает заказ-наряд вместе с позициями для страницы просмотра и PDF.
func (a *App) loadOrder(ctx context.Context, id int) (Order, []OrderItem, error) {
	var o Order
	err := a.db.QueryRow(ctx, `SELECT o.id,o.status,o.problem_description,o.manager_comment,o.total_price,o.created_at,c.full_name,cars.brand || ' ' || cars.model || ' ' || coalesce(cars.license_plate,'') FROM orders o JOIN clients c ON c.id=o.client_id JOIN cars ON cars.id=o.car_id WHERE o.id=$1`, id).Scan(&o.ID, &o.Status, &o.ProblemDescription, &o.ManagerComment, &o.TotalPrice, &o.CreatedAt, &o.ClientName, &o.CarTitle)
	if err != nil {
		return Order{}, nil, err
	}
	rows, err := a.db.Query(ctx, "SELECT id,item_type,title,quantity,price,total FROM order_items WHERE order_id=$1 ORDER BY id", id)
	if err != nil {
		return o, nil, err
	}
	defer rows.Close()
	items := []OrderItem{}
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ID, &it.Type, &it.Title, &it.Quantity, &it.Price, &it.Total); err != nil {
			return o, nil, err
		}
		items = append(items, it)
	}
	return o, items, rows.Err()
}

// getClients возвращает клиентов для выпадающих списков и таблиц.
func (a *App) getClients(ctx context.Context) []Client {
	rows, _ := a.db.Query(ctx, "SELECT id,full_name,phone,email,notes FROM clients ORDER BY full_name")
	defer rows.Close()
	items := []Client{}
	for rows.Next() {
		var c Client
		_ = rows.Scan(&c.ID, &c.FullName, &c.Phone, &c.Email, &c.Notes)
		items = append(items, c)
	}
	return items
}

// getCars возвращает автомобили для выпадающих списков и таблиц.
func (a *App) getCars(ctx context.Context) []Car {
	rows, _ := a.db.Query(ctx, `SELECT cars.id, client_id, clients.full_name, brand, model, year, vin, license_plate, mileage FROM cars JOIN clients ON clients.id=cars.client_id ORDER BY brand, model`)
	defer rows.Close()
	items := []Car{}
	for rows.Next() {
		var c Car
		_ = rows.Scan(&c.ID, &c.ClientID, &c.ClientName, &c.Brand, &c.Model, &c.Year, &c.VIN, &c.LicensePlate, &c.Mileage)
		items = append(items, c)
	}
	return items
}
