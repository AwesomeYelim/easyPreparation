document.getElementById("submitButton")?.addEventListener("click", async () => {
    const token = document.getElementById("token").value;
    const key = document.getElementById("key").value;

    if (!token || !key) {
        alert("Both token and key must be provided.");
        return;
    }

    try {
    // Go로 값을 전달 (Go에서 바인딩한 sendTokenAndKey 호출)
        await window.sendTokenAndKey(token, key);  // Go로 전달된 값

        const responseMessage = document.getElementById("responseMessage");
        responseMessage.textContent = "Success! Data sent to Go.";
    } catch (error) {
        console.error('Error:', error);
    }
});
