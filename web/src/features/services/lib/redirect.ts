// jsdom cannot follow window.location changes, so tests replace this module.
export function goTo(url: string) {
  window.location.assign(url);
}
