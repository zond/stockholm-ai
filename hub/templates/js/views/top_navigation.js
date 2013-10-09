window.TopNavigationView = Backbone.View.extend({

	template: _.template($('#top_navigation_underscore').html()),

	initialize: function() {
	  this.buttons = [
		  {
			  url: '/',
				label: 'About',
			},
			{
			  url: '/ais',
				label: 'AIs',
			},
		];
		this.activeUrl = null;
		this.listenTo(window.session.user, 'change', this.render);
	},

	navigate: function(url) {
	  window.session.router.navigate(url, { trigger: true });
		this.activeUrl = url;
		this.render();
	},

	render: function() {
		var that = this;
		that.$el.html(that.template({}));
		_.each(that.buttons, function(button) {
		  if (that.activeUrl == button.url || ('/' + that.activeUrl) == button.url) {
				that.$('ul.navbar-left').append('<li class="active"><a class="navigate" href="' + button.url + '">' + button.label + '</a></li>');
			} else {
				that.$('ul.navbar-left').append('<li><a class="navigate" href="' + button.url + '">' + button.label + '</a></li>');
			}
		});
		if (window.session.user.loggedIn()) {
		  that.$('ul.navbar-right').append('<li><a class="btn navbar-btn btn-xs" href="/logout">Sign out</a></li>');
		} else {
		  that.$('ul.navbar-right').append('<li><a class="btn navbar-btn btn-xs" href="/login">Sign in</a></li>');
		}
		$('#content').css('margin-top', that.$el.height());
		return that;
	},

});

