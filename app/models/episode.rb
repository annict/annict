class Episode < ActiveRecord::Base
  has_paper_trail

  belongs_to :work, counter_cache: true
  has_many   :checkins

  after_create :create_nicoch_program


  def number_title
    "#{number}「#{title}」"
  end


  private

  def create_nicoch_program
    if work.broadcast_on_nicoch?
      channel = Channel.find_by(name: 'ニコニコチャンネル')
      nicoch_started_day = (7 * work.episodes.count) - 7
      started_at = work.nicoch_started_at + nicoch_started_day.day

      work.programs.create(channel_id: channel.id, episode_id: id, started_at: started_at)
    end
  end
end