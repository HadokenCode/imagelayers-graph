(function() {
  'use strict';

  /**
   * @ngdoc function
   * @name iLayers.controller:BadgedialogCtrl
   * @description
   * # BadgedialogCtrl
   * Controller of the iLayers
   */
  angular.module('iLayers')
    .controller('BadgeDialogCtrl', BadgeDialogCtrl);

  BadgeDialogCtrl.$inject = ['$scope', '$sce']; 

  function BadgeDialogCtrl($scope, $sce) {
    var nameChanged = function(current, previous) {
      var newName = current.name,
          oldName = previous.name;

      return (newName !== undefined && 
          newName !== oldName);
    };

    var newImage = function() {
      return { name: '', tag: 'latest' };
    };

    $scope.selectedImage = newImage();

    $scope.imageList = function() {
      var data = $scope.graph,
          list = [];

      angular.forEach(data, function(image) {
        image.repo.label = image.repo.name + ':' + image.repo.tag;
        list.push(image.repo);
      });

      if (list.length < 1) {
        $scope.selectedWorkflow = 'hub'; 
      }

      return list;
    };

    $scope.$watch('selectedWorkflow', function() {
      $scope.selectedImage = newImage();
    });

    $scope.$watch('selectedImage', function(newValue, oldValue) {
      var image = $scope.selectedImage;

      if ($scope.selectedWorkflow === 'imagelayers' && image.name.length > 0) {
        $scope.selectedImage.selected = true; 
      }

      if ($scope.selectedWorkflow === 'hub') {
        if (image.missing || nameChanged(newValue, oldValue)) {
          $scope.selectedImage.selected = false;  
        } 
      }

      $scope.htmlCopied = false;
      $scope.markdownCopied = false;
    }, true);

    $scope.badgeAsHtml = function () {
      if ($scope.selectedImage.selected !== true) {
        return "";
      }
      
      return $sce.trustAsHtml("<a href='http://imagelayers.iron.io/?images=" + $scope.selectedImage.name + ":" + $scope.selectedImage.tag + "' title='Get your own badge on imagelayers.iron.io'>" +
      "<img src='http://badge-imagelayers.iron.io/" + $scope.selectedImage.name + ":" + $scope.selectedImage.tag + ".svg'></a>");
    };

    $scope.badgeAsMarkdown = function () {
      if ($scope.selectedImage.selected !== true) {
        return "";
      }
      
      return "[![](http://badge-imagelayers.iron.io/" + $scope.selectedImage.name + ":" + $scope.selectedImage.tag + ".svg)]" +
        "(http://imagelayers.iron.io/?images=" + $scope.selectedImage.name + ":" + $scope.selectedImage.tag + " 'Get your own badge on imagelayers.iron.io')";
    };
  }
})();