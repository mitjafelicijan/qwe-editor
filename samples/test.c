#include <stdlib.h>

// Comment 1
// Comment 2

#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/keysym.h>
#include <X11/XF86keysym.h>
#include <X11/cursorfont.h>

#include "glitch.h"
#include "config.h"

#define MAX 100

extern WindowManager wm;

static Atom _NET_WM_DESKTOP;
static Atom _NET_CURRENT_DESKTOP;
static Atom _NET_NUMBER_OF_DESKTOPS;
static Atom _NET_CLIENT_LIST;
static Atom _NET_WM_STATE;
static Atom _NET_WM_STATE_FULLSCREEN;
static Atom _NET_ACTIVE_WINDOW;

void init_window_manager(void) {
    wm.dpy = XOpenDisplay(NULL);
    if (!wm.dpy) {
        log_message(stdout, LOG_ERROR, "Cannot open display");
        abort();
    }

    wm.screen = DefaultScreen(wm.dpy);
    wm.root = RootWindow(wm.dpy, wm.screen);

    // Create and sets up cursors.
    wm.cursor_default = XCreateFontCursor(wm.dpy, XC_left_ptr);
    wm.cursor_move = XCreateFontCursor(wm.dpy, XC_fleur);
    wm.cursor_resize  = XCreateFontCursor(wm.dpy, XC_sizing);
    XDefineCursor(wm.dpy, wm.root, wm.cursor_default);
    log_message(stdout, LOG_DEBUG, "Setting up default cursors");

    // Root window input selection masks.
    XSelectInput(wm.dpy, wm.root,
            SubstructureRedirectMask | SubstructureNotifyMask |
            FocusChangeMask | EnterWindowMask | LeaveWindowMask |
            ButtonPressMask | ExposureMask | PropertyChangeMask);

    // Initialize EWMH atoms.
    _NET_WM_DESKTOP = XInternAtom(wm.dpy, "_NET_WM_DESKTOP", False);
    _NET_CURRENT_DESKTOP = XInternAtom(wm.dpy, "_NET_CURRENT_DESKTOP", False);
    _NET_NUMBER_OF_DESKTOPS = XInternAtom(wm.dpy, "_NET_NUMBER_OF_DESKTOPS", False);
    _NET_CLIENT_LIST = XInternAtom(wm.dpy, "_NET_CLIENT_LIST", False);
    _NET_WM_STATE = XInternAtom(wm.dpy, "_NET_WM_STATE", False);
    _NET_WM_STATE_FULLSCREEN = XInternAtom(wm.dpy, "_NET_WM_STATE_FULLSCREEN", False);
    _NET_ACTIVE_WINDOW = XInternAtom(wm.dpy, "_NET_ACTIVE_WINDOW", False);

    // Set number of desktops and current desktop.
    static unsigned long num_desktops = NUM_DESKTOPS;
    XChangeProperty(wm.dpy, wm.root, _NET_NUMBER_OF_DESKTOPS, XA_CARDINAL, 32, PropModeReplace, (unsigned char *)&num_desktops, 1);
    XChangeProperty(wm.dpy, wm.root, _NET_CURRENT_DESKTOP, XA_CARDINAL, 32, PropModeReplace, (unsigned char *)&num_desktops, 1);
    log_message(stdout, LOG_DEBUG, "Registering %d desktops", NUM_DESKTOPS);

    // Grab keys for keybinds.
    for (unsigned int i = 0; i < LENGTH(keybinds); i++) {
        KeyCode keycode = XKeysymToKeycode(wm.dpy, keybinds[i].keysym);
        if (keycode) {
            XGrabKey(wm.dpy, keycode, keybinds[i].mod, wm.root, True, GrabModeAsync, GrabModeAsync);
            log_message(stdout, LOG_DEBUG, "Grabbed key: mod=0x%x, keysym=0x%lx", keybinds[i].mod, keybinds[i].keysym);
        }
    }

    // Grab keys for shortcuts.
    for (unsigned int i = 0; i < LENGTH(shortcuts); i++) {
        KeyCode keycode = XKeysymToKeycode(wm.dpy, shortcuts[i].keysym);
        if (keycode) {
            XGrabKey(wm.dpy, keycode, shortcuts[i].mod, wm.root, True, GrabModeAsync, GrabModeAsync);
            log_message(stdout, LOG_DEBUG, "Grabbed shortcut: mod=0x%x, keysym=0x%lx, command=%s", shortcuts[i].mod, shortcuts[i].keysym, shortcuts[i].cmd);
        }
    }

    // Grab keys for window dragging (with MODKEY).
    XGrabButton(wm.dpy, 1, MODKEY, wm.root, True, ButtonPressMask|ButtonReleaseMask|PointerMotionMask, GrabModeAsync, GrabModeAsync, None, None);
    XGrabButton(wm.dpy, 3, MODKEY, wm.root, True, ButtonPressMask|ButtonReleaseMask|PointerMotionMask, GrabModeAsync, GrabModeAsync, None, None);
    log_message(stdout, LOG_DEBUG, "Registering grab keys for window dragging");

    // Prepare border colors.
    wm.cmap = DefaultColormap(wm.dpy, wm.screen);
    XColor active_color, inactive_color, sticky_active_color, sticky_inactive_color, dummy;

    wm.borders.normal_active = BlackPixel(wm.dpy, wm.screen);
    wm.borders.normal_inactive = BlackPixel(wm.dpy, wm.screen);
    wm.borders.sticky_active = BlackPixel(wm.dpy, wm.screen);
    wm.borders.sticky_inactive = BlackPixel(wm.dpy, wm.screen);

    if (XAllocNamedColor(wm.dpy, wm.cmap, active_border_color, &active_color, &dummy)) {
        wm.borders.normal_active = active_color.pixel;
    }

    if (XAllocNamedColor(wm.dpy, wm.cmap, inactive_border_color, &inactive_color, &dummy)) {
        wm.borders.normal_inactive = inactive_color.pixel;
    }

    if (XAllocNamedColor(wm.dpy, wm.cmap, sticky_active_border_color, &sticky_active_color, &dummy)) {
        wm.borders.sticky_active = sticky_active_color.pixel;
    }

    if (XAllocNamedColor(wm.dpy, wm.cmap, sticky_inactive_border_color, &sticky_inactive_color, &dummy)) {
        wm.borders.sticky_inactive = sticky_inactive_color.pixel;
    }
    
    XSync(wm.dpy, False);
}

void deinit_window_manager(void) {
    XFreeCursor(wm.dpy, wm.cursor_default);
    XFreeCursor(wm.dpy, wm.cursor_move);
    XFreeCursor(wm.dpy, wm.cursor_resize);
}

static int ignore_x_error(Display *dpy, XErrorEvent *err) {
    (void)dpy;
    (void)err;
    return 0;
}

int window_exists(Window window) {
    if (window == None) return 0;
    XErrorHandler old = XSetErrorHandler(ignore_x_error);
    XWindowAttributes attr;
    Status status = XGetWindowAttributes(wm.dpy, window, &attr);
    XSync(wm.dpy, False);
    XSetErrorHandler(old);
    return status != 0;
}

void set_active_window(Window window) {
    if (window != None) {
        XChangeProperty(wm.dpy, wm.root, _NET_ACTIVE_WINDOW, XA_WINDOW, 32, PropModeReplace, (unsigned char *)&window, 1);
        wm.active = window;
    } else {
        XDeleteProperty(wm.dpy, wm.root, _NET_ACTIVE_WINDOW);
    }
    XFlush(wm.dpy);
}

Window get_active_window(void) {
    Atom _NET_ACTIVE_WINDOW = XInternAtom(wm.dpy, "_NET_ACTIVE_WINDOW", False);
    Atom actual_type;
    int actual_format;
    unsigned long nitems, bytes_after;
    unsigned char *prop = NULL;
    Window active = None;

    if (XGetWindowProperty(wm.dpy, wm.root, _NET_ACTIVE_WINDOW, 0, (~0L), False, AnyPropertyType, &actual_type, &actual_format, &nitems, &bytes_after, &prop) == Success) {
        if (prop && nitems >= 1) {
            active = *(Window *)prop;
        }
    }

    if (prop) XFree(prop);
    return active;
}

void get_cursor_offset(Window window, int *dx, int *dy) {
    Window root, child;
    int root_x, root_y;
    unsigned int mask;
    XQueryPointer(wm.dpy, window, &root, &child, &root_x, &root_y, dx, dy, &mask);
}

// https://tronche.com/gui/x/xlib/events/structure-control/map.html
void handle_map_request(void) {
    Window window = wm.ev.xmaprequest.window;

    // Move window under cursor position and clamps inside the screen bounds.
    XWindowAttributes check_attr;
    if (XGetWindowAttributes(wm.dpy, window, &check_attr)) {
        XSelectInput(wm.dpy, window, EnterWindowMask | LeaveWindowMask);

        Window root_return, child_return;
        int root_x, root_y, win_x, win_y;
        unsigned int mask;

        if (XQueryPointer(wm.dpy, wm.root, &root_return, &child_return, &root_x, &root_y, &win_x, &win_y, &mask)) {
            int new_x = root_x - (check_attr.width / 2);
            int new_y = root_y - (check_attr.height / 2);
            int screen_width = DisplayWidth(wm.dpy, wm.screen);
            int screen_height = DisplayHeight(wm.dpy, wm.screen);

            if (new_x < 0) new_x = 0;
            if (new_y < 0) new_y = 0;
            if (new_x + check_attr.width > screen_width) new_x = screen_width - check_attr.width;
            if (new_y + check_attr.height > screen_height) new_y = screen_height - check_attr.height;

            XMoveWindow(wm.dpy, window, new_x, new_y);
            log_message(stdout, LOG_DEBUG, "Positioned new window 0x%lx at cursor (%d, %d)", window, root_x, root_y);
        }
    }

    // Shows, raises and focuses the window.
    set_active_border(window);
    set_active_window(window);

    XMapWindow(wm.dpy, window);
    XRaiseWindow(wm.dpy, window);
    XSetInputFocus(wm.dpy, window, RevertToPointerRoot, CurrentTime);

    log_message(stdout, LOG_DEBUG, "Window 0x%lx mapped", window);
}

// https://tronche.com/gui/x/xlib/events/window-state-change/unmap.html
void handle_unmap_notify(void) {
    Window window = wm.ev.xunmap.window;
    log_message(stdout, LOG_DEBUG, "Window 0x%lx unmapped", window);
}

// https://tronche.com/gui/x/xlib/events/window-state-change/destroy.html
void handle_destroy_notify(void) {
    Window window = wm.ev.xdestroywindow.window;
    log_message(stdout, LOG_DEBUG, "Window 0x%lx destroyed", window);
}

// https://tronche.com/gui/x/xlib/events/client-communication/property.html
void handle_property_notify(void) {
    Window window = wm.ev.xproperty.window;
    Atom prop = wm.ev.xproperty.atom;
    char *name = XGetAtomName(wm.dpy, prop);
    log_message(stdout, LOG_DEBUG, "Window 0x%lx got property notification %s", window, name);
}

// https://tronche.com/gui/x/xlib/events/keyboard-pointer/keyboard-pointer.html
void handle_motion_notify(void) {
    if (wm.start.subwindow != None && (wm.start.state & MODKEY)) {
        int xdiff = wm.ev.xmotion.x_root - wm.start.x_root;
        int ydiff = wm.ev.xmotion.y_root - wm.start.y_root;

        XMoveResizeWindow(wm.dpy, wm.start.subwindow,
                wm.attr.x + (wm.start.button == 1 ? xdiff : 0),
                wm.attr.y + (wm.start.button == 1 ? ydiff : 0),
                MAX(100, wm.attr.width  + (wm.start.button == 3 ? xdiff : 0)),
                MAX(100, wm.attr.height + (wm.start.button == 3 ? ydiff : 0)));
    }
}

// https://tronche.com/gui/x/xlib/events/client-communication/client-message.html
void handle_client_message(void) {
    Window window = wm.ev.xclient.window;
    int message_type = wm.ev.xclient.message_type;
    log_message(stdout, LOG_DEBUG, "Window 0x%lx got message type of %d", window, message_type);
}

// https://tronche.com/gui/x/xlib/events/keyboard-pointer/keyboard-pointer.html
void handle_button_press(void) {
    Window window = wm.ev.xbutton.subwindow;
    if (window == None) return;

    if (wm.ev.xbutton.state & MODKEY) {
        XRaiseWindow(wm.dpy, window);
        XGetWindowAttributes(wm.dpy, window, &wm.attr);
        wm.start = wm.ev.xbutton;

        set_active_border(window);
        set_active_window(window);

        switch (wm.ev.xbutton.button) {
            case 1: {
                XDefineCursor(wm.dpy, window, wm.cursor_move);
                log_message(stdout, LOG_DEBUG, "Setting cursor to move");
            } break;
            case 3: {
                XDefineCursor(wm.dpy, window, wm.cursor_resize);
                log_message(stdout, LOG_DEBUG, "Setting cursor to resize");
            } break;
        }

        log_message(stdout, LOG_DEBUG, "Window 0x%lx got button press press", window);
        XFlush(wm.dpy);
    }
}

// https://tronche.com/gui/x/xlib/events/keyboard-pointer/keyboard-pointer.html
void handle_button_release(void) {
    Window window = wm.ev.xbutton.subwindow;
    if (window == None) return;

    if (wm.start.state & MODKEY) {
        XDefineCursor(wm.dpy, wm.start.subwindow, None);
        log_message(stdout, LOG_DEBUG, "Setting cursor to resize");
    }

    log_message(stdout, LOG_DEBUG, "Window 0x%lx got button release", window);
    XFlush(wm.dpy);
}

// https://tronche.com/gui/x/xlib/events/keyboard-pointer/keyboard-pointer.html
void handle_key_press(void) {
    log_message(stdout, LOG_DEBUG, ">> Key pressed > active window 0x%lx", wm.ev.xkey.subwindow);
    if (wm.ev.type != KeyPress) return;
    if (wm.ev.xkey.subwindow == None) return;

    // TODO: Check why XkbKeycodeToKeysym worked in previous version.
    /* KeySym keysym = XkbKeycodeToKeysym(wm.dpy, wm.ev.xkey.keycode, 0, 0); */
    KeySym keysym = XLookupKeysym(&wm.ev.xkey, 0);

    // Check keybinds first.
    for (unsigned int i = 0; i < LENGTH(keybinds); i++) {
        if (keysym == keybinds[i].keysym && (wm.ev.xkey.state & (Mod1Mask|Mod2Mask|Mod3Mask|Mod4Mask|ControlMask|ShiftMask)) == keybinds[i].mod) {
            keybinds[i].func(&keybinds[i].arg);
            break;
        }
    }

    XFlush(wm.dpy);
}

void handle_key_release(void) {}

void handle_focus_in(void) {
    Window window = wm.ev.xfocus.window;
    if (window != wm.root) {
        log_message(stdout, LOG_DEBUG, "Window 0x%lx focus in", window);
    }
}

void handle_focus_out(void) {
    Window window = wm.ev.xfocus.window;
    if (window != wm.root) {
        log_message(stdout, LOG_DEBUG, "Window 0x%lx focus out", window);
    }
}

void handle_enter_notify(void) {
    Window window = wm.ev.xcrossing.window;
    if (window != wm.root) {
        set_active_border(window);
        set_active_window(window);
        log_message(stdout, LOG_DEBUG, "Window 0x%lx enter notify", window);
    }
}

void set_active_border(Window window) {
    if (window == None) return;

    // Setting current active window to inactive.
    if (wm.active != None) {
        XSetWindowBorderWidth(wm.dpy, wm.active, border_size);
        XSetWindowBorder(wm.dpy, wm.active, wm.borders.normal_inactive);
        log_message(stdout, LOG_DEBUG, "Active window 0x%lx border set to inactive", window);
    }

    // Setting desired window to active.
    XSetWindowBorderWidth(wm.dpy, window, border_size);
    XSetWindowBorder(wm.dpy, window, wm.borders.normal_active);
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Desired window 0x%lx border set to active", window);
}

void move_window_x(const Arg *arg) {
    if (wm.active == None) return;

    XWindowAttributes attr;
    XGetWindowAttributes(wm.dpy, wm.active, &attr);
    XMoveWindow(wm.dpy, wm.active, attr.x + arg->i, attr.y);
    log_message(stdout, LOG_DEBUG, "Move window 0x%lx on X by %d", wm.active, arg->i);

    int rel_x, rel_y;
    get_cursor_offset(wm.active, &rel_x, &rel_y);
    XWarpPointer(wm.dpy, None, wm.active, 0, 0, 0, 0, rel_x + arg->i, rel_y);

    XSync(wm.dpy, True);
    XFlush(wm.dpy);
}

void move_window_y(const Arg *arg) {
    if (wm.active == None) return;

    XWindowAttributes attr;
    XGetWindowAttributes(wm.dpy, wm.active, &attr);
    XMoveWindow(wm.dpy, wm.active, attr.x, attr.y + arg->i);
    log_message(stdout, LOG_DEBUG, "Move window 0x%lx on Y by %d", wm.active, arg->i);

    int rel_x, rel_y;
    get_cursor_offset(wm.active, &rel_x, &rel_y);
    XWarpPointer(wm.dpy, None, wm.active, 0, 0, 0, 0, rel_x, rel_y + arg->i);

    XSync(wm.dpy, True);
    XFlush(wm.dpy);
}

void resize_window_x(const Arg *arg) {
    if (wm.active == None) return;

    XWindowAttributes attr;
    XGetWindowAttributes(wm.dpy, wm.active, &attr);
    XResizeWindow(wm.dpy, wm.active, MAX(1, attr.width + arg->i), attr.height);
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Resize window 0x%lx on X by %d", wm.active, arg->i);
}

void resize_window_y(const Arg *arg) {
    if (wm.active == None) return;

    XWindowAttributes attr;
    XGetWindowAttributes(wm.dpy, wm.active, &attr);
    XResizeWindow(wm.dpy, wm.active, attr.width, MAX(1, attr.height + arg->i));
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Resize window 0x%lx on Y by %d", wm.active, arg->i);
}

void window_snap_up(const Arg *arg) {
    (void)arg;
    if (wm.active == None) return;

    XWindowAttributes attr;
    if (!XGetWindowAttributes(wm.dpy, wm.active, &attr)) {
        log_message(stdout, LOG_DEBUG, "Failed to get window attributes for 0x%lx", wm.active);
        return;
    }

    int rel_x, rel_y;
    get_cursor_offset(wm.active, &rel_x, &rel_y);

    XMoveWindow(wm.dpy, wm.active, attr.x, 0);
    XWarpPointer(wm.dpy, None, wm.active, 0, 0, 0, 0, rel_x, rel_y);
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Snapped window 0x%lx to top edge", wm.active);
}

void window_snap_down(const Arg *arg) {
    (void)arg;
    if (wm.active == None) return;

    XWindowAttributes attr;
    if (!XGetWindowAttributes(wm.dpy, wm.active, &attr)) {
        log_message(stdout, LOG_DEBUG, "Failed to get window attributes for 0x%lx", wm.active);
        return;
    }

    int rel_x, rel_y;
    int y = DisplayHeight(wm.dpy, DefaultScreen(wm.dpy)) - attr.height - (2 * attr.border_width);
    get_cursor_offset(wm.active, &rel_x, &rel_y);
    
    XMoveWindow(wm.dpy, wm.active, attr.x, y);
    XWarpPointer(wm.dpy, None, wm.active, 0, 0, 0, 0, rel_x, rel_y);
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Snapped window 0x%lx to bottom edge", wm.active);
}

void window_snap_left(const Arg *arg) {
    (void)arg;
    if (wm.active == None) return;

    XWindowAttributes attr;
    if (!XGetWindowAttributes(wm.dpy, wm.active, &attr)) {
        log_message(stdout, LOG_DEBUG, "Failed to get window attributes for 0x%lx", wm.active);
        return;
    }

    int rel_x, rel_y;
    get_cursor_offset(wm.active, &rel_x, &rel_y);
    
    XMoveWindow(wm.dpy, wm.active, 0, attr.y);
    XWarpPointer(wm.dpy, None, wm.active, 0, 0, 0, 0, rel_x, rel_y);
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Snapped window 0x%lx to left edge", wm.active);
}

void window_snap_right(const Arg *arg) {
    (void)arg;
    if (wm.active == None) return;

    XWindowAttributes attr;
    if (!XGetWindowAttributes(wm.dpy, wm.active, &attr)) {
        log_message(stdout, LOG_DEBUG, "Failed to get window attributes for 0x%lx", wm.active);
        return;
    }

    int rel_x, rel_y;
    int x = DisplayWidth(wm.dpy, DefaultScreen(wm.dpy)) - attr.width - (2 * attr.border_width);
    get_cursor_offset(wm.active, &rel_x, &rel_y);
    
    XMoveWindow(wm.dpy, wm.active, x, attr.y);
    XWarpPointer(wm.dpy, None, wm.active, 0, 0, 0, 0, rel_x, rel_y);
    XFlush(wm.dpy);

    log_message(stdout, LOG_DEBUG, "Snapped window 0x%lx to right edge", wm.active);
}
