var persona = angular.module("persona", []);

persona.factory('sessionService', ["$http", "$q", function($http, $q) {
    var _sessionKey = 'mtrack.session';
    var _cached;
    var _session = function() {
		if (typeof _cached !== 'undefined') {
			// TODO check expiration time.
			return Object.create(_cached);
		}

		var cached = localStorage[_sessionKey];
		if (typeof cached !== 'undefined') { 
			_cached = JSON.parse(cached);
			return Object.create(_cached);
		}

		return { status: 'unauthenticated' };
    };
    var _store = function(session) {
		if (typeof session === 'object') {
			_cached = session;
			localStorage[_sessionKey] = JSON.stringify(session);
			console.log('local storage success');
			return true;
		} else {
			console.log('local storage failure');
			return false;
		}
    };
    var _delete = function() {
		if (typeof _cached !== 'undefined') {
			_cached = undefined;
			delete localStorage[_sessionKey];
			return true;
		}
		return false;
    };
    return {
        session: _session,
        store: _store,
        endSession: _delete
    };
}]);

persona.factory("personaService", ["$http", "$q", "sessionService", function($http, $q, sessionService) {
    return {
        verify: function () {
                    var deferred = $q.defer();
                    navigator.id.get(function(assertion) {
                        $http.post(config.persona.verifyUrl, { assertion: assertion }).
                            success(function(data) {
                                console.log('verify success:', data);
                                sessionService.store(data);
                                deferred.resolve(data);
                            }).
                            error(function(data) {
                                console.log('verify failure:', data);
                                deferred.reject(data.reason);
                            });
                    });
                    return deferred.promise;
                },
        logout: function () {
                    var deferred = $q.defer();
                    if (config.persona.logoutUrl) {

                        $http.post(config.persona.logoutUrl).
                            success(function(data) {
                                sessionService.endSession();
								deferred.resolve(true);
                            }).
                            error(function(data) { deferred.resolve(true); });
                    } else {
                        sessionService.endSession();
						deferred.resolve(true);
                    }
                    return deferred.promise;
                }
    };
}]);
