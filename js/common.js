$(function(){
  var zero = function(s, n){
    s = '0000' + s;
    return s.substring(s.length - n);
  };
  $('.create_at').each(function(){
    var timestamp = $(this).text();
    var d = new Date();
    d.setTime(timestamp.substring(0, timestamp.length - 3));
    $(this).text(
      zero(d.getFullYear(), 4) + '-' +
      zero((d.getMonth() + 1), 2) + '-' +
      zero(d.getDate(), 2) + ' ' +
      zero(d.getHours(), 2) + ':' +
      zero(d.getMinutes(), 2) + ':' +
      zero(d.getSeconds(), 2));
    $(this).show();
  });
  $('a.subject_link').click(function(){
    document.location.href = '/s/' + encodeURIComponent($(this).text().substring(1));
    return false;
  });
  $('div.text').each(function(){
    var html = $(this).html();
    var url_regexp = /(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/g;
    html = html.replace(url_regexp, function(x){
      return '<a class="twitter-url" href="' + x + '" target="_blank">' + x + '</a>';
    });
    var hashtag_regexp = /[#＃][^ .;:　]+/g;
    html = html.replace(hashtag_regexp, function(x){
      return '<a class="twitter-hashtag" href="http://twitter.com/#!/search?q=' + encodeURIComponent(x) + '" target="_blank">' + x + '</a>';
    });
    var at_regexp = /@[_a-z0-9]+/ig;
    html = html.replace(at_regexp, function(x){
      return '<a class="twitter-at" href="http://twitter.com/#!/' + encodeURIComponent(x.substring(1)) + '" target="_blank">' + x + '</a>';
    });
    $(this).html(html);
  });
});
//  vim: set ts=2 sw=2 sts=2 expandtab fenc=utf-8:
