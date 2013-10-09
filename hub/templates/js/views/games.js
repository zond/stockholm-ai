window.GamesView = Backbone.View.extend({
	
	template: _.template($('#games_underscore').html()),

	initialize: function() {
	  this.collection = new Games();
		this.listenTo(this.collection, 'reset', this.render);
		this.listenTo(this.collection, 'add', this.render);
		this.listenTo(this.collection, 'remove', this.render);
		this.collection.fetch({ reset: true });
		this.ais = new AIs();
		this.listenTo(this.ais, 'reset', this.render);
		this.listenTo(this.ais, 'add', this.render);
		this.listenTo(this.ais, 'remove', this.render);
		this.ais.fetch({ reset: true });
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({}));
		that.collection.each(function(game) {
			that.$('table').append('<tr><td>' + game.get('Players') + '</td></tr>');
		});
		that.ais.each(function(ai) {
      that.$('select').append('<option value="' + ai.get('Id') + '">' + ai.get('Name') + '</option>');
		});
		if (window.session.user.loggedIn()) {
		  that.$('.add-game').show();
		} else {
		  that.$('.add-game').hide();
		}
		that.$('.multiselect').multiselect();
		return that;
	},

});
