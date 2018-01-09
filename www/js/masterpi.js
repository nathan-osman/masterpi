/**
 * Masterpi JavaScript
 * Copyright 2018 - Nathan Osman
 */

$(function() {

    // Find the page components
    var $group = $('.group.lamp'),
        $statusVal = $group.find('.status .value'),
        $on = $group.find('button.on'),
        $off = $group.find('button.off');

    /**
     * Load the current state of the lamp
     */
    function loadState() {
        return $.get('/api/lamp/state')
        .done(function(d) {
            $statusVal.text(d.value ? "on" : "off");
        });
    }

    /**
     * Set the state of the lamp
     */
    function setState(value) {
        return $.ajax({
            type: 'POST',
            url: '/api/lamp/state',
            data: JSON.stringify({
                value: value
            }),
            contentType: 'application/json'
        });
    }

    /**
     * Handle an error from an AJAX request
     */
    function handleError(promise) {
        promise.fail(function() {
            $statusVal.text('error');
        });
    }

    // Set the click handlers for the "on" and "off" buttons
    $on.click(function() {
        handleError(setState(true).then(loadState));
    });
    $off.click(function() {
        handleError(setState(false).then(loadState));
    });

    // Load the initial value
    handleError(loadState());

});
