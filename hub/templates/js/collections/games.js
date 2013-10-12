window.Games = Backbone.Collection.extend({

	model: Game,

	url: function() {
	  return '/games?offset=' + ((this.page - 1) * this.limit) + '&limit=' + this.limit;
	},

	initialize: function(options) {
	  this.page = options.page || 1;
		this.limit = options.limit || 10;
		this.pages = 0;
	},

	parse: function(data) {
	  this.pages = Math.ceil(data.Total / this.limit);
	  return data.Content;
	},

});


