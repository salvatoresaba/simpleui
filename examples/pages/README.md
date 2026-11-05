# Pages example

This example shows how a SimpleUI project can contain **multiple pages** and how
to move between them at runtime.

Run it with:

```bash
simpleui serve examples/pages/app.sui --watch
```

Or build it:

```bash
simpleui build examples/pages/app.sui --out dist
simpleui build examples/pages/app.sui --single --out dist/pages.html
```

## The page model

Every non-invoked root `view` is a page:

```sui
view main {
	title "Pages"
	h1 "Home"
}

view projects {
	title "Projects"
	h1 "Projects"
}

view settings {
	title "Settings"
	h1 "Settings"
}
```

The generated application keeps all three pages in the DOM. Only the **first**
page is visible when the app starts; the others are hidden. This means:

- switching pages does not reload the browser;
- state declared at the top level is shared by every page;
- the `title` of the visible page becomes the browser tab title;
- page content is already rendered when it is shown.

The first page in the source is the initial page. Reordering the views changes
the initial page.

## Switching pages

Use the built-in function:

```sui
navigate("pageName")
```

The name must match the `view` name. `navigate` does not require an import and
returns `null`. It is a no-op for unknown names and for single-page apps.

A typical handler is:

```sui
fn openSettings() -> null {
	visits = visits + 1
	navigate("settings")
}
```

Then wire it to a button or menu item:

```sui
button "Settings" onClick: openSettings
```

```sui
menuitem "Settings" onClick: openSettings
```

## Shared state

The example declares one state at project level:

```sui
state visits = 0
```

Every page displays it. Because state is global, the count survives page
switches:

```sui
p "Navigations: {visits}"
```

Use page-specific state only when needed; it still lives in the same runtime
scope unless you use component instance state.

## Using the designer

1. Open the project in the designer.
2. Use the **page chips above the preview** to switch which page you are
   editing.
3. **Insert → New page…** creates another page.
4. The **View** inspector section edits the active page title/name and offers
   **Duplicate**, **Delete** and **Save as template**.
5. Add buttons or menu items to each page; their handlers call `navigate`.
6. In the preview, page chips switch the editor view. The generated app uses
   the handlers, not the chips.

Only one page can be edited or previewed at a time in the designer. That is an
editor restriction; the generated app keeps all pages available.

## Current limitations

`navigate` is internal navigation. It does not currently provide:

- URL routes or deep links;
- browser back/forward integration;
- page transition animations;
- automatic navigation bars.

Build the navigation UI explicitly with buttons, menus or tabs. A future
`pagenav` component could generate the links automatically if needed.

## Related documentation

- [API reference: pages and navigation](../../docs/API_REFERENCE.md#pages-and-navigation)
- [GUI editor: pages](../../docs/GUI_EDITOR.md#pages)
- [Builtins: navigation](../../docs/BUILTINS.md#navigation)
- [Pages example source](app.sui)
