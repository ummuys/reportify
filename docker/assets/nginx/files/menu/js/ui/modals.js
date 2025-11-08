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


export function showConfirm(message, title = "Подтверждение", optionYes = 'Удалить') {
    return new Promise(resolve => {
        const modal = document.getElementById('confirmModal');
        const msgEl = document.getElementById('confirmMessage');
        const btnYes = document.getElementById('confirmYes');
        const btnNo = document.getElementById('confirmNo');

        document.querySelector('.modal-title').textContent = title;
        btnYes.textContent = optionYes;
        msgEl.textContent = message;
        modal.style.display = 'flex';

        const close = (result) => {
            modal.style.display = 'none';
            btnYes.removeEventListener('click', yesHandler);
            btnNo.removeEventListener('click', noHandler);
            document.removeEventListener('keydown', keyHandler); // Удаляем обработчик клавиш
            resolve(result);
        };

        const yesHandler = () => close(true);
        const noHandler = () => close(false);

        const keyHandler = (event) => {
            if (event.key === 'Enter') {
                event.preventDefault();
                yesHandler();
            } else if (event.key === 'Escape') {
                event.preventDefault();
                noHandler();
            }
        };

        btnYes.addEventListener('click', yesHandler);
        btnNo.addEventListener('click', noHandler);
        document.addEventListener('keydown', keyHandler); // Добавляем обработчик клавиш
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