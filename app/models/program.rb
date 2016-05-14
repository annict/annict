# == Schema Information
#
# Table name: programs
#
#  id             :integer          not null, primary key
#  channel_id     :integer          not null
#  episode_id     :integer          not null
#  work_id        :integer          not null
#  started_at     :datetime         not null
#  created_at     :datetime
#  updated_at     :datetime
#  sc_last_update :datetime
#  sc_pid         :integer
#  rebroadcast    :boolean          default(FALSE), not null
#
# Indexes
#
#  index_programs_on_sc_pid  (sc_pid) UNIQUE
#

class Program < ActiveRecord::Base
  include DbActivityMethods
  include ProgramCommon

  has_paper_trail only: DIFF_FIELDS

  has_many :draft_programs, dependent: :destroy

  scope :episode_published, -> { joins(:episode).merge(Episode.published) }
  scope :work_published, -> { joins(:work).merge(Work.published) }
end
