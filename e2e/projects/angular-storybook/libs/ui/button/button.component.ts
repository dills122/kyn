import { Component, Input } from '@angular/core';

@Component({
  selector: 'ui-button',
  templateUrl: './button.component.html',
})
export class ButtonComponent {
  @Input() label = 'Click me';
}
