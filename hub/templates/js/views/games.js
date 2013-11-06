window.GamesView = Backbone.View.extend({
	
	template: _.template($('#games_underscore').html()),

	events: {
	  "click .create-button": 'createGame',
		"click .go-to-page": 'goToPage',
		"click .last-page": 'lastPage',
		"click .first-page": 'firstPage',
	},

	initialize: function() {
	  this.collection = new Games([], {});
		this.listenTo(this.collection, 'sync', this.render);
		this.listenTo(this.collection, 'reset', this.render);
		this.listenTo(this.collection, 'add', this.render);
		this.listenTo(this.collection, 'remove', this.render);
		this.collection.fetch({ reset: true });
		this.ais = new AIs();
		this.listenTo(this.ais, 'sync', this.render);
		this.listenTo(this.ais, 'reset', this.render);
		this.listenTo(this.ais, 'add', this.render);
		this.listenTo(this.ais, 'remove', this.render);
		this.ais.fetch({ reset: true });
	},

	firstPage: function(ev) {
	  ev.preventDefault();
		this.collection.page = 1;
		this.collection.fetch({ reset: true });
	},

	lastPage: function(ev) {
	  ev.preventDefault();
		this.collection.page = this.collection.pages;
		this.collection.fetch({ reset: true });
	},

  goToPage: function(ev) {
	  ev.preventDefault();
		this.collection.page = $(ev.target).attr('data-page');
		this.collection.fetch({ reset: true });
	},

	createGame: function(ev) {
		var that = this;
	  ev.preventDefault();
		that.collection.create({
		  Players: that.$('select').val(),
			State: 'Created',
			Length: 0,
			PlayerNames: _.collect(that.$('select').val(), function(id) {
			  return that.ais.get(id).get('Name')
			}),
		}, { at: 0 });
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({
		  page: that.collection.page,
			limit: that.collection.limit,
			pages: that.collection.pages,
		}));
		that.collection.each(function(game) {
			that.$('table').append(new ListedGameView({
			  model: game,
				collection: that.collection,
			}).render().el);
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
