# == Schema Information
#
# Table name: episodes
#
#  id             :integer          not null, primary key
#  work_id        :integer          not null
#  number         :string(510)
#  sort_number    :integer          default("0"), not null
#  sc_count       :integer
#  title          :string(510)
#  checkins_count :integer          default("0"), not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  episodes_work_id_idx           (work_id)
#  episodes_work_id_sc_count_key  (work_id,sc_count) UNIQUE
#

class Episode < ActiveRecord::Base
  has_paper_trail

  belongs_to :work, counter_cache: true
  has_many :checkins
  has_many :programs

  after_create :create_nicoch_program


  def number_title
    "#{number}「#{title}」"
  end

  # 映画やOVAなどの実質エピソードを持たない作品かどうかを判定する
  def single?
    number.blank? && title.present?
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
