import Cookies from 'js-cookie'

export const zoom = {
    state: Cookies.get('zoom') || 'medium',
    transitions: {
        medium: {
            in() {
                this.state = 'large'
            },
            out() {
                this.state = 'small'
            }
        },
        small: {
            in() {
                this.state = 'medium'
            },
            out() {
                this.state = 'xs'
            }
        },
        large: {
            in() {
                this.state = 'xl'
            },
            out() {
                this.state = 'medium'
            }
        },
        xl: {
            out() {
                this.state = 'large'
            },
        },
        xs: {
            in() {
                this.state = 'small'
            }
        }
    },
    dispatch(actionName) {
        const oldState = this.state
        const action = this.transitions[this.state][actionName];
        if (action) {
            action.call(this);
            Cookies.set('zoom', this.state);
            document.body.classList.replace(`zoom-${oldState}`, `zoom-${this.state}`)
        }
    },
};