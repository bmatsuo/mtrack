mtrack.filter('moment', function() {
    return function(timestamp, directive) {
        if (directive === 'ago') {
            return moment(timestamp).fromNow();
        } else { // assume a format string in all other cases.
            return moment(timestamp).format(directive);
        }
    };
});
