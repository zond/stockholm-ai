window.AIErrorsView = Backbone.View.extend({
	
	template: _.template($('#ai_errors_underscore').html()),

	collapseTemplate: _.template($('#collapsible_underscore').html()),

	initialize: function(options) {
	  this.collection = new AIErrors([], {
		  url: '/ais/' + options.id + '/errors',
		});
		this.listenTo(this.collection, 'sync', this.render);
		this.listenTo(this.collection, 'reset', this.render);
		this.listenTo(this.collection, 'add', this.render);
		this.listenTo(this.collection, 'remove', this.render);
		this.collection.fetch({ reset: true });
	},

  render: function() {
		var that = this;
    that.$el.html(that.template({}));
		that.collection.each(function(err) {
		  that.$('#accordion').append(that.collapseTemplate({
			  title: err.get('CreatedAt') + ' ' + err.get('Error'),
				body: '<pre>' + err.get('ErrorDetail1') + '</pre><pre>' + err.get('ErrorDetail2') + '</pre>',				
			}));
		});
		return that;
	},

});
