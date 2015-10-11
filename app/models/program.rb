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
#
# Indexes
#
#  programs_channel_id_episode_id_key  (channel_id,episode_id) UNIQUE
#  programs_channel_id_idx             (channel_id)
#  programs_episode_id_idx             (episode_id)
#  programs_work_id_idx                (work_id)
#

class Program < ActiveRecord::Base
  include DbActivityMethods
  include ProgramCommon

  has_paper_trail only: DIFF_FIELDS

  belongs_to :work
  has_many :draft_programs, dependent: :destroy

  scope :work_published, -> { joins(:work).merge(Work.published) }
end
