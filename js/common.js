$(function(){
  var zero = function(s, n){
    s = '0000' + s;
    return s.substring(s.length - n);
  };

  //initialize
  var initialize = function(target){
    $('.create_at', target).each(function(){
      var timestamp = $(this).text();
      var d = new Date();
      d.setTime(timestamp.substring(0, timestamp.length - 3));
      var date_string = 
        zero(d.getFullYear(), 4) + '年' +
        zero((d.getMonth() + 1), 2) + '月' +
        zero(d.getDate(), 2) + '日 ' +
        zero(d.getHours(), 2) + ':' +
        zero(d.getMinutes(), 2);
      $(this).html(
        '<a target="_blank" href="http://twitter.com/#!/' +
        $(this).data('screenname') + '/status/' +
        $(this).data('statusid') + '">' +
        date_string + '</a>');
      $(this).show();
    });
    $('a.subject_link', target).click(function(){
      document.location.href = '/s/' + encodeURIComponent($(this).text().substring(1));
      return false;
    });
    $('div.text', target).each(function(){
      var html = $(this).html();
      var url_regexp = /(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/g;
      html = html.replace(url_regexp, function(x){
        return '<a class="twitter-url" href="' + x + '" target="_blank">' + x + '</a>';
      });
      var hashtag_regexp = /[#＃][^ .;:　\n]+/g;
      html = html.replace(hashtag_regexp, function(x){
        return '<a class="twitter-hashtag" href="http://twitter.com/#!/search?q=' + encodeURIComponent(x) + '" target="_blank">' + x + '</a>';
      });
      var at_regexp = /@[_a-z0-9]+/ig;
      html = html.replace(at_regexp, function(x){
        return '<a class="twitter-at" href="http://twitter.com/#!/' + encodeURIComponent(x.substring(1)) + '" target="_blank">' + x + '</a>';
      });
      $(this).html(html);
    });
    $('div.twitter-buttons div.button', target).each(function(){
      var $this = $(this);
      if ($this.hasClass('retweeted') || $this.hasClass('favorited')) return;
      $this.parent().attr('href', $this.data('href'));
    });
    //events
    $('a.profile,a.reply,a.retweet,a.favorited', target).click(function(){
      var $this = $(this);
      var type = $this.hasClass('profile') ? 'profile' :
                 $this.hasClass('reply') ? 'reply' :
                 $this.hasClass('favorite') ? 'favorite' :
                 $this.hasClass('retweet') ? 'retweet' :
                 '';
      // increment point.
      if (!type) return false;
      $.post('/point_up', {type: type, key: $this.data('key')},
        function(data){
        }
      );
      return false;
    });
    $('div.like a.like', target).click(function(){
      var $this = $(this);
      $.post('/like', {key: $this.data('key'), url: document.location.pathname},
        function(data){
          if (data == 'needs_oauth') {
            document.location.href = '/get_request_token';
            return;
          }
          // update page.
          if (data != '') {
            var image = document.createElement('img');
            image.src = 'http://img.tweetimag.es/i/' + data + '_m';
            $('.users', $this.parent().parent().parent().parent()).append(image);
          }
        }
      );
      $('body').css('cursor', 'default');
      return false;
    });
  };
  initialize(document);

  $('div.more-tweets').click(function(){
    var $this = $(this);
    var hashtag = $this.data('hashtag');
    var page = $this.data('page') + 1;
    $.get('/s/more', {hashtag: hashtag, page: page}, function(data){
      if (!data) {
        $('div.more-tweets').hide('slow');
        return;
      }
      var data_element = $(data);
      initialize(data_element);
      $('div#subject_blocks').append(data_element);
      $this.data('page', page);
    });
  });

  //ticker
  var ticker_width =
    $('h1').css('width').replace('px', '') -
    $('h1 a').css('width').replace('px', '') - 10;
  $('div#ticker').css('width', ticker_width + 'px');
  $('div#ticker-elements').html($('div#ticker-elements-hide').html());
  $('div#ticker-elements').jStockTicker({interval: 13});
});

var _gaq = _gaq || [];
_gaq.push(['_setAccount', 'UA-25384750-1']);
_gaq.push(['_trackPageview']);

(function() {
  var ga = document.createElement('script'); ga.type = 'text/javascript'; ga.async = true;
  ga.src = ('https:' == document.location.protocol ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js';
  var s = document.getElementsByTagName('script')[0]; s.parentNode.insertBefore(ga, s);
})();
//woopra
function woopraReady(tracker) {
  tracker.setDomain('e-hash.jp');
  tracker.setIdleTimeout(300000);
  tracker.track();
  return false;
}
(function() {
  var wsc = document.createElement('script');
  wsc.src = document.location.protocol+'//static.woopra.com/js/woopra.js';
  wsc.type = 'text/javascript';
  wsc.async = true;
  var ssc = document.getElementsByTagName('script')[0];
  ssc.parentNode.insertBefore(wsc, ssc);
})();

//  vim: set ts=2 sw=2 sts=2 expandtab fenc=utf-8:
