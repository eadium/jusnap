define([
    "require",
    "jquery",
    "base/js/namespace",
], function (require, $, Jupyter) {
    "use strict";

    var mod_name = 'Snapshot_menu';
    var mod_log_prefix = mod_name + '[' + mod_name + ']';


    var options = {
        sibling: undefined, // if undefined, set by cfg.sibling_selector
        menus : [],
        hooks: {
            pre_config: undefined,
            post_config: undefined,
        }
    };

    var cfg = {
        api_host: 'localhost:8000',
        snapshots_submenu_id: '#snapshots-menu-load',
        insert_before_sibling: false,
        sibling_selector: '#help_menu',
        top_level_submenu_goes_left: true,
    };

    function notifyMe(msg, options) {
        if (!("Notification" in window)) {
          alert("This browser does not support desktop notification");
        }
        else if (Notification.permission === "granted") {
          var notification = new Notification(msg, options);
        }
        else if (Notification.permission !== "denied") {
          Notification.requestPermission().then(function (permission) {
            if (permission === "granted") {
              var notification = new Notification(msg, options);
            } else {
                alert('Please allow notifications to control snapshotting')
            }
          });
        }
      }
      

    async function http_get_async(theUrl, callback) {
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => { 
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                callback(xmlHttp.responseText);
            } else if (xmlHttp.status != 200 && xmlHttp.status != 0) {
                console.log("GET req " + theUrl + ": " + xmlHttp.status + ", " + xmlHttp.responseText)
            }
        }
        xmlHttp.open("GET", theUrl, true);
        xmlHttp.send(null);
    }

    async function http_post_async(theUrl, body, callback) {
        var xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => { 
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                    callback(xmlHttp.responseText);
            } else if (xmlHttp.status != 200 && xmlHttp.status != 0) {
                console.log("POST req " + theUrl + ": " + xmlHttp.status + ", " + xmlHttp.responseText)
            }
            
        }
        xmlHttp.open("POST", theUrl, true);
        xmlHttp.send(JSON.stringify(body));
    }

    function async_load_snapshots(resp) {
        var snapshots = JSON.parse(resp)['snapshots']
        var el = document.querySelector(cfg.snapshots_submenu_id)
        snapshots.forEach(e => {

            let new_el = {
                'name': e['date'],
                'load_snapshot': e['id']
            }
            let submenu = build_menu_element(new_el, 'right')
            submenu.appendTo(el)
        })

    }

    function load_snapshots() {
        http_get_async('http://'+cfg.api_host+'/api/snap', async_load_snapshots)
    }

    function create_snapshot() {
        http_post_async('http://'+cfg.api_host+'/api/snap/new', {}, (resp) => {
            var data = JSON.parse(resp)
            $(cfg.snapshots_submenu_id).empty()
            load_snapshots()
            var message
            if (data['status'] === 'snapshotted') {
                message = 'Snapshot created'
            } else if (data['status'] === 'skipped') {
                message = 'Snapshotting skipped: cooldown'
            } else {
                message = data
            }

            notifyMe(message)
        })
    }

    function callback_create_snapshot(evt) {
        create_snapshot()
    }

    function load_snapshot(id) {
        var req = {
            'id': id.toString()
        }
        http_post_async('http://'+cfg.api_host+'/api/snap/restore', req, (resp) => {
            var data = JSON.parse(resp)
            var message 
            var opts = {}
            if ('id' in data) {
                message = 'Loaded snapshot'
                opts['body'] = format_time(id)
            } else if ('msg' in data) {
                message = 'Snapshot warning'
                opts['body'] = data['msg']
            } else {
                message = 'Snapshot warning'
                opts['body'] = resp
            }

            notifyMe(message, opts)
        })
    }

    function callback_load_snapshot(evt) {
        load_snapshot($(evt.currentTarget).data('snapshot-id'))
    }

    function config_loaded_callback () {
        if (options['pre_config_hook'] !== undefined) {
            options['pre_config_hook']();
        }

        load_snapshots()

        options.menus = [
            {
                'name' : 'Snapshots',
                'sub-menu-direction' : 'right',
                'sub-menu' : [
                    {
                        'name' : 'Create',
                        'sub-menu-direction' : 'right',
                        'create_snapshot': ''
                    },
                    {
                        'name' : 'Load',
                        'sub-menu-direction' : 'right',
                        'id': 'snapshots-menu-load',
                        'sub-menu': []
                    }
                ],
            },
        ];

        if (options.hooks.post_config !== undefined) {
            options.hooks.post_config();
        }

        // select correct sibling
        if (options.sibling === undefined) {
            options.sibling = $(cfg.sibling_selector).parent();
            if (options.sibling.length < 1) {
                options.sibling = $("#help_menu").parent();
            }
        }
    }

    function build_menu_element (menu_item_spec, direction) {
        // Create the menu item html element
        var element = $('<li/>');

        if (typeof menu_item_spec == 'string') {
            if (menu_item_spec != '---') {
                console.log(mod_log_prefix,
                    'Don\'t understand sub-menu string "' + menu_item_spec + '"');
                return null;
            }
            return element.addClass('divider');
        }

        var a = $('<a/>')
            .attr('href', '#')
            .html(menu_item_spec.name)
            .appendTo(element);
        
        if (menu_item_spec.hasOwnProperty('create_snapshot')) {
            a.attr({
                'title' : "",
            })
            .on('click', callback_create_snapshot)
            .addClass('snapshot');
        }
        if (menu_item_spec.hasOwnProperty('load_snapshot')) {
            var snapshot = menu_item_spec.load_snapshot;
            if (typeof snapshot == 'string' || snapshot instanceof String) {
                snapshot = [snapshot];
            }
            a.attr({
                'title' : "",
                'data-snapshot-id' : snapshot,
            })
            .on('click', callback_load_snapshot)
            .addClass('snapshot');
        }
        

        if (menu_item_spec.hasOwnProperty('sub-menu')) {
            element
                .addClass('dropdown-submenu')
                .toggleClass('dropdown-submenu-left', direction === 'left');
            var sub_element = $('<ul class="dropdown-menu"/>')
                .toggleClass('dropdown-menu-compact', menu_item_spec.overlay === true) // For space-saving menus
                .appendTo(element);
            
            if (menu_item_spec.hasOwnProperty('id') && menu_item_spec.id === 'snapshots-menu-load') {
                sub_element.attr({
                    'id': menu_item_spec.id
                })
            }

            var new_direction = (menu_item_spec['sub-menu-direction'] === 'left') ? 'left' : 'right';
            for (var j=0; j<menu_item_spec['sub-menu'].length; ++j) {
                var sub_menu_item_spec = build_menu_element(menu_item_spec['sub-menu'][j], new_direction);
                if(sub_menu_item_spec !== null) {
                    sub_menu_item_spec.appendTo(sub_element);
                }
            }
        }

        return element;
    }

    function menu_setup (menu_item_specs, sibling, insert_before_sibling) {
        for (var i=0; i<menu_item_specs.length; ++i) {
            var menu_item_spec;
            if (insert_before_sibling) {
                menu_item_spec = menu_item_specs[i];
            } else {
                menu_item_spec = menu_item_specs[menu_item_specs.length-1-i];
            }
            var direction = (menu_item_spec['menu-direction'] == 'left') ? 'left' : 'right';
            var menu_element = build_menu_element(menu_item_spec, direction);
            // We need special properties if this item is in the navbar
            if ($(sibling).parent().is('ul.nav.navbar-nav')) {
                menu_element
                    .addClass('dropdown')
                    .removeClass('dropdown-submenu dropdown-submenu-left');
                menu_element.children('a')
                    .addClass('dropdown-toggle')
                    .attr({
                        'data-toggle' : 'dropdown',
                        'aria-expanded' : 'false',
                        'id': 'snapshots-menu'
                    });
            }

            // Insert the menu element into DOM
            menu_element[insert_before_sibling ? 'insertBefore': 'insertAfter'](sibling);

            // Make sure MathJax will typeset this menu
            window.MathJax.Hub.Queue(["Typeset", window.MathJax.Hub, menu_element[0]]);
        }
    }

    function format_time(ts) {
        const milliseconds = ts * 1000 
        const dateObject = new Date(milliseconds)
        const humanDateFormat = dateObject.toLocaleString() 
        return humanDateFormat
    }

    function load_ipython_extension () {
        Jupyter.notebook.config.loaded.then(
            config_loaded_callback
        ).then(function () {
            menu_setup(options.menus, options.sibling, cfg.insert_before_sibling);
        });
    }

    return {
        load_ipython_extension : load_ipython_extension,
        menu_setup : menu_setup,
        options : options,
    };

});