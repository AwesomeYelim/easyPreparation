document.addEventListener("DOMContentLoaded", () => {
    const greeting = document.getElementById("greeting");
    const button = document.getElementById("updateButton");

    button.addEventListener("click", () => {
        greeting.textContent = "You clicked the button!";
    });
});
