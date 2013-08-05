mtrack.controller('ProgressCtrl', ["$scope", "$http", "$q", "personaService", "sessionService",function ProgressCtrl($scope, $http, $q, personaService, sessionService) {
    $scope.verified = false;
    $scope.userId = undefined;
    $scope.media = {};
    $scope.usersInProgress = {};
    $scope.usersFinished = {};
    $scope.mediaRoots = [];
    $scope.mediaByRoot = {};
    var mediaByRoot = {},
        mediaById = {},
        users = [],
        userById = {};

    var logApiError = function(data, status) {
        console.log("HTTP status", status.code, data.reason);
    };

    $scope.verify = function() {
        personaService.verify().
            then(function(result) {
                console.log('verified', result);
                $scope.verified = true;
                $scope.userId = result.userId;
            }, function(result) {
                console.log('verification failure', result);
            });
    };

    $scope.logout = function() {
        personaService.logout().
            then(function(result) {
                console.log('logged out', result);
                $scope.verified = false;
                $scope.userId = undefined;
            });
    };

    $scope.getProgress = function() {
        var resp = $http.get('/api/media/progress');
        resp.success(function(data, status, headers) {
            $scope.usersInProgress = {};
            $scope.usersFinished = {};
            for (i in $scope.media) {
                var m = $scope.media[i].mediaId;
                $scope.usersInProgress[m.mediaId] = [];
                $scope.usersFinished[m.mediaId] = [];
            }
            var progress = data.results;
            for (i in progress) {
                var p = progress[i];
                var users = [];
                if (typeof p.finished !== 'undefined') {
                    users = $scope.usersFinished[p.mediaId] || [];
                    if (typeof users === 'undefined') users = [];
                    users = users.concat(p);
                    $scope.usersFinished[p.mediaId] = users;
                } else {
                    users = $scope.usersInProgress[p.mediaId] || [];
                    if (typeof users === 'undefined') users = [];
                    users = users.concat(p);
                    $scope.usersInProgress[p.mediaId] = users;
                }
            }
        });
        resp.error(function(data, status, headers) {
            logApiError(data, status);
        });
    };

    $scope.getMedia = function() {
        var resp = $http.get('/api/media');
        resp.success(function(data, status, headers) {
            $scope.mediaByRoot = {};
            var media = data.results;

            mediaByRoot = {};
            mediaById = {};

            for (i in media) {
                var m = media[i],
                    rootMedia = $scope.mediaByRoot[m.root],
                    path = m.path,
                    base = new String(path).substring(path.lastIndexOf('/') + 1); 

                m.basename = base;
                media[i] = m;
                mediaById[m.mediaId] = m;
                if (typeof rootMedia === 'undefined') rootMedia = [];
                $scope.mediaByRoot[m.root] = rootMedia.concat(m);
            }

            $scope.mediaRoots = [];
            for (root in $scope.mediaByRoot) {
                $scope.mediaRoots = $scope.mediaRoots.concat(root);
            }

            $scope.media = media;
        });
        resp.error(function(data, status, headers) {
            logApiError(data, status);
        });
    };

    $scope.mediaUnwatched = function(mediaId) {
        return !$scope.mediaStarted(mediaId) && !$scope.mediaFinished(mediaId);
    };

    $scope.mediaStarted = function(mediaId) {
        var users = $scope.usersInProgress[mediaId] || [];
        for (i in users) {
            if (users[i].userId == $scope.userId) {
                console.log('finished', mediaId);
                return true;
            }
        }
        console.log('unfinished', mediaId);
        return false;
    };

    $scope.mediaFinished = function(mediaId) {
        var users = $scope.usersFinished[mediaId] || [];
        for (i in users) {
            if (users[i].userId == $scope.userId) {
                console.log('finished', mediaId);
                return true;
            }
        }
        console.log('unfinished', mediaId);
        return false;
    };

    $scope.startMedia = function(mediaId) {
        var session = sessionService.session();
        var accessToken = session.accessToken;
        var userId = session.userId;
        var data = { userId: userId, mediaId: mediaId };
        var header = { Authorization: 'token ' + accessToken};
        $http.post('/api/start', data, { headers: header }).
            success(function(data) {
                $scope.getProgress();
            }).
        error(function(data, status) {
            console.log('startMedia:', status.code, data);
        });
    };

    $scope.finishMedia = function(mediaId) {
        var sessino = sessionService.session();
        var accessToken = session.accessToken;
        var userId = session.userId;
        var data = { userId: userId, mediaId: mediaId };
        var header = { Authorization: 'token ' + accessToken};
        $http.post('/api/finish', data, { headers: header }).
            success(function(data) {
                console.log('finish that shit');
                $scope.getProgress();
            }).
        error(function(data, status) {
            console.log('startMedia:', status.code, data);
        });
    };

    $scope.clearMedia = function(mediaId) {
        var sessino = sessionService.session();
        var accessToken = session.accessToken;
        var userId = session.userId;
        var data = { userId: userId, mediaId: mediaId };
        var header = { Authorization: 'token ' + accessToken};
        $http.post('/api/clear', data, { headers: header }).
            success(function(data) {
                console.log('finish that shit');
                $scope.getProgress();
            }).
        error(function(data, status) {
            console.log('startMedia:', status.code, data);
        });
    };

    var session = sessionService.session();
    if (typeof session.accessToken !== 'undefined') {
        $scope.verified = true;
        $scope.userId = session.userId;
    }
    $scope.getMedia();
    $scope.getProgress();
}]);
