import type { Meta, StoryObj } from '@storybook/angular';
import { ButtonComponent } from './button.component';

const meta: Meta<ButtonComponent> = {
  title: 'UI/Button',
  component: ButtonComponent,
};
export default meta;

type Story = StoryObj<ButtonComponent>;

export const Primary: Story = {
  args: { label: 'Primary' },
};
