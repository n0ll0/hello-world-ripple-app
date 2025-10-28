// @ts-expect-error: known issue, we're working on it
import { mount } from 'ripple';
import { Root } from './Root.ripple';

mount(Root, {
	target: document.getElementById('root'),
});
