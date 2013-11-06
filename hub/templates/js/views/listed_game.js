window.ListedGameView = Backbone.View.extend({

	template: _.template($('#listed_game_underscore').html()),

	tagName: 'tr',

	events: {
	  'click .clone-button': 'cloneGame',
	},

	initialize: function() {
	},

	cloneGame: function(ev) {
	  {{if .DevServer}}
		var that = this;
	  ev.preventDefault();
		that.collection.create({
		  Players: that.model.get('Players'),
			State: 'Created',
			Length: 0,
			PlayerNames: that.model.get('PlayerNames'),
		}, { at: 0 });
		{{end}}
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({
		  model: that.model,
		}));
		return that;
	},

});

