/**
 * Masterpi JavaScript
 * Copyright 2018 - Nathan Osman
 */

$(function() {

    function toList(value) {
        return $.map(value.split(','), function(v) {
            return v.trim();
        });
    }

    // Find the page components
    var $lampGroup = $('.group.lamp'),
        $lampStatus = $lampGroup.find('.status .value'),
        $lampOn = $lampGroup.find('button.on'),
        $lampOff = $lampGroup.find('button.off'),
        $timerGroup = $('.group.timer'),
        $timerOn = $timerGroup.find('.turn-on'),
        $timerOff = $timerGroup.find('.turn-off'),
        $timerSave = $('.timer button');

    /**
     * Load the current state of the lamp
     */
    function loadState() {
        return $.get('/api/lamp/state')
        .done(function(d) {
            $lampStatus.text(d.value ? "on" : "off");
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
     * Load the values for the timers
     */
    function loadTimerValues() {
        return $.get('/api/timer/values')
        .done(function(d) {
            $timerOn.val(d['turn-on'].join(', '));
            $timerOff.val(d['turn-off'].join(', '));
        });
    }

    /**
     * Save the timer values
     */
    function saveTimerValues() {
        return $.ajax({
            type: 'POST',
            url: '/api/timer/values',
            data: JSON.stringify({
                'turn-on': toList($timerOn.val()),
                'turn-off': toList($timerOff.val())
            }),
            contentType: 'application/json'
        });
    }

    /**
     * Handle an error from an AJAX request
     */
    function handleError(promise) {
        promise.fail(function() {
            $lampStatus.text('error');
        });
    }

    // Set the click handlers for the "on" and "off" buttons
    $lampOn.click(function() {
        handleError(setState(true).then(loadState));
    });
    $lampOff.click(function() {
        handleError(setState(false).then(loadState));
    });

    // Set the click handler for the timer
    $timerSave.click(function() {
        handleError(saveTimerValues());
    });

    // Load the initial values
    handleError(loadState());
    handleError(loadTimerValues());

});
