class Loader {
	constructor() {
		this.mask = null
		this.init()
	}

	init() {
		// Создаем маску если её нет
		if (!this.mask) {
			this.mask = document.createElement('div')
			this.mask.className = 'loader-mask'
			this.mask.innerHTML = '<div class="loader"></div>'
			document.body.appendChild(this.mask)
		}
	}

	show() {
		this.init()
		this.mask.classList.remove('hide')
	}

	hide() {
		if (this.mask) {
			this.mask.classList.add('hide')
			// Удаляем элемент после анимации
			setTimeout(() => {
				if (this.mask && this.mask.parentNode) {
					this.mask.parentNode.removeChild(this.mask)
					this.mask = null
				}
			}, 600)
		}
	}

	//Показать на время загрузки страницы
	showOnLoad() {
		this.show()
		window.addEventListener('load', () => {
			setTimeout(() => this.hide(), 500)
		})
	}
}

// Создаем глобальный экземпляр
export const loader = new Loader();
