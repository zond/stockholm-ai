window.AIsView = Backbone.View.extend({
	
	template: _.template($('#ais_underscore').html()),

	events: {
	  'click .add-ai button': 'createNewAI',
	},

	initialize: function() {
	  this.collection = new AIs();
		this.listenTo(this.collection, 'reset', this.render);
		this.listenTo(this.collection, 'add', this.render);
		this.listenTo(this.collection, 'remove', this.render);
		this.collection.fetch({ reset: true });
	},

	createNewAI: function(ev) {
	  ev.preventDefault();
		this.collection.create({
		  Name: $('.new-ai-name').val(),
			URL: $('.new-ai-url').val(),
			Owner: window.session.user.get('Email'),
		});
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({}));
		that.collection.each(function(ai) {
		  that.$('table').append('<tr><td>' + ai.get('Name') + '</td><td>' + ai.get('URL') + '</td><td>' + ai.get('Owner') + '</td></tr>');
		});
		if (window.session.user.loggedIn()) {
		  that.$('.add-ai').show();
		} else {
		  that.$('.add-ai').hide();
		}
		return that;
	},

});
