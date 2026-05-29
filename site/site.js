const copyButtons = document.querySelectorAll('[data-copy]');

for (const button of copyButtons) {
  const originalLabel = button.textContent;
  button.addEventListener('click', async () => {
    const command = button.getAttribute('data-copy');
    if (!command) {
      return;
    }

    await navigator.clipboard.writeText(command);
    button.textContent = 'Copied';
    window.setTimeout(() => {
      button.textContent = originalLabel;
    }, 1800);
  });
}
