import { mount } from 'svelte';
import App from './App.svelte';

const alvo = document.getElementById('app');
if (!alvo) throw new Error('elemento #app não encontrado');

export default mount(App, { target: alvo });
