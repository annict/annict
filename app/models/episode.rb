# == Schema Information
#
# Table name: episodes
#
#  id             :integer          not null, primary key
#  work_id        :integer          not null
#  number         :string(255)
#  sort_number    :integer          default(0), not null
#  title          :string(255)
#  created_at     :datetime
#  updated_at     :datetime
#  checkins_count :integer          default(0), not null
#  single         :boolean          default(FALSE)
#  sc_count       :integer
#
# Indexes
#
#  index_episodes_on_checkins_count        (checkins_count)
#  index_episodes_on_work_id_and_sc_count  (work_id,sc_count) UNIQUE
#

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
