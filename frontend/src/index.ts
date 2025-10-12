import { mount } from 'ripple';
// @ts-expect-error: known issue, we're working on it
import { Root } from './Root.ripple';

mount(Root, {
	target: document.getElementById('root'),
});
