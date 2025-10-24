export function showToast(message, duration = 3000) {
    let toastTimeout;
    const toast = document.getElementById('toast');
    if (!toast) {
        console.warn('Toast element not found');
        return;
    }

    clearTimeout(toastTimeout);
    toast.classList.remove('show');

    toast.textContent = message;

    void toast.offsetWidth;

    toast.classList.add('show');

    toastTimeout = setTimeout(() => {
        toast.classList.remove('show');
    }, duration);
}


export function showConfirm(message, title = "Подтверждение") {
    return new Promise(resolve => {
        const modal = document.getElementById('confirmModal');
        const msgEl = document.getElementById('confirmMessage');
        const btnYes = document.getElementById('confirmYes');
        const btnNo = document.getElementById('confirmNo');

        document.querySelector('.modal-title').textContent = title;
        msgEl.textContent = message;
        modal.style.display = 'flex';

        const close = (result) => {
        modal.style.display = 'none';
        btnYes.removeEventListener('click', yesHandler);
        btnNo.removeEventListener('click', noHandler);
        resolve(result);
        };

        const yesHandler = () => close(true);
        const noHandler = () => close(false);

        btnYes.addEventListener('click', yesHandler);
        btnNo.addEventListener('click', noHandler);
    });
}


export function showAlert(message, title = "Сообщение") {
    return new Promise(resolve => {
        const modal = document.getElementById('alertModal');
        const msgEl = document.getElementById('alertMessage');
        const titleEl = document.getElementById('alertTitle');
        const btnOk = document.getElementById('alertOk');

        msgEl.textContent = message;
        titleEl.textContent = title;
        modal.style.display = 'flex';

        const close = () => {
        modal.style.display = 'none';
        btnOk.removeEventListener('click', close);
        resolve();
        };

        btnOk.addEventListener('click', close);
    });
}