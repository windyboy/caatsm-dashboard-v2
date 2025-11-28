// Design System Usage Examples
// This file demonstrates common composition patterns and usage examples

import type { ComponentSize, ComponentVariant } from '../../types/ui';

/**
 * Example: Form with consistent styling
 */
export const formExample = `
<!-- Login Form Example -->
<Card variant="default" size="md" class="max-w-md mx-auto">
  <form class="space-y-4">
    <div>
      <label for="email" class="block text-sm font-medium text-slate-700 mb-1">
        Email Address
      </label>
      <Input
        id="email"
        type="email"
        placeholder="Enter your email"
        variant="primary"
        size="md"
        required
      />
    </div>

    <div>
      <label for="password" class="block text-sm font-medium text-slate-700 mb-1">
        Password
      </label>
      <Input
        id="password"
        type="password"
        placeholder="Enter your password"
        variant="primary"
        size="md"
        required
      />
    </div>

    <div class="flex gap-3">
      <Button variant="primary" size="md" type="submit" class="flex-1">
        Sign In
      </Button>
      <Button variant="outline" size="md" type="button">
        Cancel
      </Button>
    </div>
  </form>
</Card>
`;

/**
 * Example: Data table with actions
 */
export const tableExample = `
<!-- Data Table Example -->
<Card variant="elevated" size="lg">
  <div class="flex justify-between items-center mb-4">
    <h3 class="text-lg font-semibold text-slate-900">Users</h3>
    <Button variant="primary" size="sm">
      Add User
    </Button>
  </div>

  <div class="overflow-x-auto">
    <table class="w-full">
      <thead class="border-b border-slate-200">
        <tr>
          <th class="text-left py-2 px-4 font-medium text-slate-700">Name</th>
          <th class="text-left py-2 px-4 font-medium text-slate-700">Email</th>
          <th class="text-left py-2 px-4 font-medium text-slate-700">Role</th>
          <th class="text-left py-2 px-4 font-medium text-slate-700">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each users as user}
          <tr class="border-b border-slate-100 hover:bg-slate-50">
            <td class="py-3 px-4 text-slate-900">{user.name}</td>
            <td class="py-3 px-4 text-slate-600">{user.email}</td>
            <td class="py-3 px-4">
              <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium
                {user.role === 'admin' ? 'bg-danger-100 text-danger-800' : 'bg-success-100 text-success-800'}">
                {user.role}
              </span>
            </td>
            <td class="py-3 px-4">
              <div class="flex gap-2">
                <Button variant="ghost" size="sm">Edit</Button>
                <Button variant="danger" size="sm">Delete</Button>
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</Card>
`;

/**
 * Example: Loading and error states
 */
export const asyncExample = `
<!-- Async Operation Example -->
<script>
  let loading = false;
  let error = null;
  let data = null;

  async function loadData() {
    loading = true;
    error = null;
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 2000));
      data = { message: 'Data loaded successfully!' };
    } catch (err) {
      error = err;
    } finally {
      loading = false;
    }
  }
</script>

<ErrorWrapper {error} showError={!!error}>
  <LoadingWrapper {loading} loadingText="Loading data...">
    <Card variant="default" size="md">
      {#if data}
        <p class="text-slate-700">{data.message}</p>
        <Button variant="primary" size="sm" on:click={loadData}>
          Reload
        </Button>
      {:else}
        <Button variant="primary" on:click={loadData}>
          Load Data
        </Button>
      {/if}
    </Card>
  </LoadingWrapper>
</ErrorWrapper>
`;

/**
 * Example: Modal with form
 */
export const modalExample = `
<!-- Modal Example -->
<script>
  let showModal = false;

  function openModal() {
    showModal = true;
  }

  function closeModal() {
    showModal = false;
  }
</script>

{#snippet footerButtons()}
  <div class="flex justify-end gap-3">
    <Button variant="outline" on:click={closeModal}>
      Cancel
    </Button>
    <Button variant="primary" type="submit">
      Create Item
    </Button>
  </div>
{/snippet}

<Button variant="primary" onclick={openModal}>
  Open Modal
</Button>

<Modal open={showModal} onclose={closeModal} title="Create New Item" size="md" footer={true} footerSnippet={footerButtons}>
  <form class="space-y-4">
    <Input
      label="Name"
      placeholder="Enter item name"
      required
    />

    <Select
      label="Category"
      placeholder="Select category"
      options={[
        { value: 'electronics', label: 'Electronics' },
        { value: 'clothing', label: 'Clothing' },
        { value: 'books', label: 'Books' }
      ]}
      required
    />
  </form>
</Modal>
`;

/**
 * Example: Toast notifications
 */
export const toastExample = `
<!-- Toast Example -->
<script>
  import { toast } from '$lib/stores/toast';

  function showSuccessToast() {
    toast.show({
      variant: 'success',
      title: 'Success!',
      message: 'Your changes have been saved.',
      duration: 3000
    });
  }

  function showErrorToast() {
    toast.show({
      variant: 'danger',
      title: 'Error',
      message: 'Failed to save changes. Please try again.',
      duration: 5000
    });
  }
</script>

<div class="space-x-4">
  <Button variant="success" on:click={showSuccessToast}>
    Show Success Toast
  </Button>
  <Button variant="danger" on:click={showErrorToast}>
    Show Error Toast
  </Button>
</div>

<!-- Toast container would be rendered globally -->
`;

/**
 * Common size and variant combinations
 */
export const commonPatterns = {
  // Button patterns
  primaryActions: { variant: 'primary' as ComponentVariant, size: 'md' as ComponentSize },
  secondaryActions: { variant: 'secondary' as ComponentVariant, size: 'md' as ComponentSize },
  destructiveActions: { variant: 'danger' as ComponentVariant, size: 'md' as ComponentSize },

  // Form patterns
  formInputs: { variant: 'primary' as ComponentVariant, size: 'md' as ComponentSize },
  formLabels: 'block text-sm font-medium text-slate-700 mb-1',

  // Card patterns
  contentCards: { variant: 'default' as const, size: 'md' as ComponentSize },
  elevatedCards: { variant: 'elevated' as const, size: 'md' as ComponentSize },
  borderedCards: { variant: 'bordered' as const, size: 'md' as ComponentSize }
};

/**
 * Accessibility guidelines
 */
export const accessibilityGuidelines = {
  buttons: [
    'Always provide meaningful text or aria-label',
    'Use appropriate variants for different actions (primary for main actions)',
    'Ensure sufficient color contrast',
    'Support keyboard navigation'
  ],
  forms: [
    'Associate labels with inputs using for/id attributes',
    'Provide error messages that are programmatically associated',
    'Use appropriate input types for better UX',
    'Ensure form validation feedback is clear'
  ],
  modals: [
    'Focus should move to modal on open',
    'Focus should return to trigger on close',
    'Escape key should close modal',
    'Backdrop click should close modal (configurable)'
  ]
};