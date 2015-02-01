var dashboard = angular.module('dashboard', [
    'ui.bootstrap',
    'dashboardControllers'
]);

var dashboardControllers = angular.module('dashboardControllers', []);

// main controller
dashboardControllers.controller('dashboardCtrl', ['$scope', '$http', '$location', '$anchorScroll', '$sce',
    function($scope, $http, $location, $anchorScroll, $sce) {

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

        $scope.LoadEnv = function(callback) {
            $http.get('/api/env').success(function(data) {
                $scope.Env = data;

                var found = false;
                for (var i in data) {
                    if (data[i].Key == "VCAP_APPLICATION") {
                        $scope.CloudFoundry = $sce.trustAsHtml($scope.SyntaxHighlight(JSON.parse(data[i].Value)));
                        found = true;
                    }
                }

                // development fallback
                /*
                if (!found) {
                    $scope.CloudFoundry = $sce.trustAsHtml($scope.SyntaxHighlight({
                        "limits": {
                            "mem": 16,
                            "disk": 1024,
                            "fds": 16384
                        },
                        "application_version": "f66320cb-1149-4746-9caf-35df5cc81ab8",
                        "application_name": "dashboard",
                        "application_uris": ["dashboard.10.244.0.34.xip.io"],
                        "version": "f66320cb-1149-4746-9caf-35df5cc81ab8",
                        "name": "dashboard",
                        "space_name": "development",
                        "space_id": "bd20fde8-64c0-4901-ad85-7f1e88214a77",
                        "uris": ["dashboard.10.244.0.34.xip.io"],
                        "users": null,
                        "application_id": "c27e4367-b522-4721-a4ea-1e6e770c9093",
                        "instance_id": "1f13709b817d496fb32f6a43a76aa285",
                        "instance_index": 0,
                        "host": "0.0.0.0",
                        "port": 61005,
                        "started_at": "2015-02-01 08:21:42 +0000",
                        "started_at_timestamp": 1422778902,
                        "start": "2015-02-01 08:21:42 +0000",
                        "state_timestamp": 1422778902
                    }));
                }
                */

                if (callback) {
                    callback();
                }
            });
        };

        $scope.SyntaxHighlight = function(json) {
            json = JSON.stringify(json, undefined, 4);
            json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
            return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function(match) {
                var cls = 'number';
                if (/^"/.test(match)) {
                    if (/:$/.test(match)) {
                        cls = 'key';
                    } else {
                        cls = 'string';
                    }
                } else if (/true|false/.test(match)) {
                    cls = 'boolean';
                } else if (/null/.test(match)) {
                    cls = 'null';
                }
                return '<span class="' + cls + '">' + match + '</span>';
            });
        }

        $scope.LoadHeaders = function(callback) {
            $http.get('/api/headers').success(function(data) {
                $scope.Headers = data;
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
            $scope.LoadEnv();
            $scope.LoadHeaders();

            if (callback) {
                callback();
            }
        };

        $scope.LoadAllData(function() {
            $location.path("/");
        });
    }
]);