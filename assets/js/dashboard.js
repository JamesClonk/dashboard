/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

var dashboard = angular.module('dashboard', [
	'ui.bootstrap',
	'dashboardControllers'
]);

var dashboardControllers = angular.module('dashboardControllers', []);

// main controller
dashboardControllers.controller('dashboardCtrl', ['$scope', '$http', '$location', '$anchorScroll',
	function($scope, $http, $location, $anchorScroll) {

		$scope.Goto = function(path) {
			$location.path(path);
		};

		$scope.ScrollTo = function(id) {
			$location.hash(id);
			$anchorScroll();
		}

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

		$scope.LoadMemory = function(callback) {
			$http.get('/api/mem').success(function(data) {
				$scope.Memory = data;
				for (var i in $scope.Memory) {
					$scope.Memory[i].UsedPercentage = Math.round(($scope.Memory[i].UsedM / $scope.Memory[i].TotalM) * 100);
					$scope.Memory[i].FreePercentage = Math.round(($scope.Memory[i].FreeM / $scope.Memory[i].TotalM) * 100);
					$scope.Memory[i].Class = "progress-bar-success";
					if ($scope.Memory[i].UsedPercentage > 85) {
						$scope.Memory[i].Class = "progress-bar-danger";
					} else if ($scope.Memory[i].UsedPercentage > 65) {
						$scope.Memory[i].Class = "progress-bar-warning";
					} else if ($scope.Memory[i].UsedPercentage > 45) {
						$scope.Memory[i].Class = "progress-bar-primary";
					}
				}

				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadDisk = function(callback) {
			$http.get('/api/disk').success(function(data) {
				$scope.Disk = data;
				for (var i in $scope.Disk) {
					$scope.Disk[i].Class = "progress-bar-success";
					if ($scope.Disk[i].UsagePercentage > 85) {
						$scope.Disk[i].Class = "progress-bar-danger";
					} else if ($scope.Disk[i].UsagePercentage > 65) {
						$scope.Disk[i].Class = "progress-bar-warning";
					} else if ($scope.Disk[i].UsagePercentage > 45) {
						$scope.Disk[i].Class = "progress-bar-primary";
					}
				}

				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadUsers = function(callback) {
			$http.get('/api/users').success(function(data) {
				$scope.Users = data;
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadLoggedOn = function(callback) {
			$http.get('/api/logged_on').success(function(data) {
				$scope.LoggedOn = data;
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadProcesses = function(callback) {
			$http.get('/api/processes').success(function(data) {
				$scope.Processes = data;
				$scope.reverse = true;
				$scope.SortField = "Cpu";
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadNetwork = function(callback) {
			$http.get('/api/network').success(function(data) {
				$scope.Network = data;
				if (callback) {
					callback();
				}
			});
		};

		$scope.LoadAllData = function(callback) {
			$scope.LoadHostname();
			$scope.LoadIP();
			$scope.LoadCPU();
			$scope.LoadMemory();
			$scope.LoadDisk();
			$scope.LoadUsers();
			$scope.LoadLoggedOn();
			$scope.LoadProcesses();
			$scope.LoadNetwork();

			if (callback) {
				callback();
			}
		};

		$scope.LoadAllData(function() {
			$location.path("/");
		});
	}
]);