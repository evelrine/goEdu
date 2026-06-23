// Фильтрация автомобилей при создании заказ-наряда.
// Автомобиль должен принадлежать выбранному клиенту.
document.addEventListener("DOMContentLoaded", function () {
  const clientSelect = document.getElementById("order-client");
  const carSelect = document.getElementById("order-car");
  const orderForm = document.getElementById("order-form");

  if (!clientSelect || !carSelect) {
    return;
  }

  const carOptions = Array.from(carSelect.options).filter(function (option) {
    return option.dataset.clientId;
  });

  function filterCars() {
    const clientId = clientSelect.value;
    const selectedCarClientId = carSelect.selectedOptions[0]?.dataset.clientId;

    carOptions.forEach(function (option) {
      option.hidden = option.dataset.clientId !== clientId;
    });

    if (!clientId) {
      carSelect.value = "";
      carSelect.options[0].textContent = "Сначала выберите клиента";
      return;
    }

    carSelect.options[0].textContent = "Выберите автомобиль";

    if (selectedCarClientId !== clientId) {
      carSelect.value = "";
    }
  }

  clientSelect.addEventListener("change", filterCars);
  filterCars();

  if (orderForm) {
    orderForm.addEventListener("submit", function (event) {
      const selectedCar = carSelect.selectedOptions[0];
      if (!selectedCar || selectedCar.dataset.clientId !== clientSelect.value) {
        event.preventDefault();
        alert("Выберите автомобиль, который принадлежит выбранному клиенту.");
      }
    });
  }
});
