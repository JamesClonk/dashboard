/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

var dashboard = angular.module('dashboard', [
	'ui.bootstrap',
	'dashboardControllers'
]);

var dashboardControllers = angular.module('dashboardControllers', []);

// main controller
dashboardControllers.controller('dashboardCtrl', ['$scope', '$http', '$location',
	function($scope, $http, $location) {

		$scope.Goto = function(path) {
			$location.path(path);
		};

		$scope.LoadHostname = function(callback) {
			$http.get('/api/hostname').success(function(data) {
				$scope.Hostname = data.Hostname;
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadIP = function(callback) {
			$http.get('/api/ip').success(function(data) {
				$scope.IP = "0.0.0.0";
				for (var ip in data) {
					if (data[ip] != null && data[ip] != "") {
						$scope.IP = data[ip];
					}
				}
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadCPU = function(callback) {
			$http.get('/api/cpu').success(function(data) {
				$scope.CPU = data;
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadAllData = function(callback) {
			$scope.LoadHostname();
			$scope.LoadIP();
			$scope.LoadCPU();

			if (callback) {
				callback();
			}
		};

		$scope.LoadAllData(function() {
			$location.path("/");
		});
	}
]);