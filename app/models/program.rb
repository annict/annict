# == Schema Information
#
# Table name: programs
#
#  id             :integer          not null, primary key
#  channel_id     :integer          not null
#  episode_id     :integer          not null
#  work_id        :integer          not null
#  started_at     :datetime         not null
#  sc_last_update :datetime
#  created_at     :datetime
#  updated_at     :datetime
#  sc_pid         :integer
#  rebroadcast    :boolean          default(FALSE), not null
#
# Indexes
#
#  index_programs_on_sc_pid  (sc_pid) UNIQUE
#  programs_channel_id_idx   (channel_id)
#  programs_episode_id_idx   (episode_id)
#  programs_work_id_idx      (work_id)
#

class Program < ActiveRecord::Base
  include DbActivityMethods
  include ProgramCommon

  has_many :draft_programs, dependent: :destroy

  scope :episode_published, -> { joins(:episode).merge(Episode.published) }
  scope :work_published, -> { joins(:work).merge(Work.published) }

  def broadcasted?(time = Time.now.in_time_zone("Asia/Tokyo"))
    time > started_at.in_time_zone("Asia/Tokyo")
  end
end
