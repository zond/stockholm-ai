window.session = {};

$(window).load(function() {
  
	$(document).on('a.navigate', function(ev) {
	  ev.preventDefault();
		window.session.nav.navigate($(ev.currentTarget).attr('href'));
	});

	var AppRouter = Backbone.Router.extend({

		routes: {
			"": "about",
		},

		about: function(id) {
			new AboutView({
				el: $('#content'),
			}).render();
		},
	});	
	
	window.session.nav = new TopNavigationView({
		el: $('nav'),
	}).render();

	window.session.router = new AppRouter();
	Backbone.history.start({ 
		pushState: true,
	});

	window.session.nav.navigate(Backbone.history.fragment || '/');
});

