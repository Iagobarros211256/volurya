/* Galeria de vídeos da Volurya. Apenas UI; nenhum endpoint é chamado. */
document.addEventListener('DOMContentLoaded', () => {
  const dialog = document.getElementById('vol-gallery-video-dialog');
  const player = document.getElementById('vol-gallery-video-player');
  const close = dialog?.querySelector('.vol-gallery-dialog-close');
  if (!dialog || !player || !close) return;

  function releaseMedia() {
    player.pause();
    player.removeAttribute('src');
    player.removeAttribute('poster');
    player.load();
  }

  document.querySelectorAll('[data-vol-gallery-video]').forEach(link => {
    link.addEventListener('click', event => {
      // Sem suporte a <dialog>, o link continua abrindo o MP4 normalmente.
      if (typeof dialog.showModal !== 'function') return;
      event.preventDefault();
      releaseMedia();
      player.poster = link.dataset.poster || '';
      player.src = link.getAttribute('href');
      dialog.showModal();
      player.play().catch(() => { /* Usuário pode apertar play manualmente. */ });
    });
  });

  close.addEventListener('click', () => dialog.close());
  dialog.addEventListener('close', releaseMedia);
  dialog.addEventListener('click', event => {
    if (event.target === dialog) dialog.close();
  });
});
