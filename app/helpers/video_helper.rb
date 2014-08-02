module VideoHelper
  def background_video_tag
    host = 'http://d3a8d1smk6xli.cloudfront.net/videos/'
    videos = [
      ['P1080730.webm', 'P1080730.ogv', 'P1080730.mp4'],
      ['P1080741.webm', 'P1080741.ogv', 'P1080741.mp4'],
      ['P1080749.webm', 'P1080749.ogv', 'P1080749.mp4'],
      ['P1080770.webm', 'P1080770.ogv', 'P1080770.mp4']
    ]

    video_tag(videos.sample.map { |video| host + video }, autoplay: true, muted: true, loop: true)
  end
end