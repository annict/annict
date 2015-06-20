# == Schema Information
#
# Table name: draft_programs
#
#  id         :integer          not null, primary key
#  program_id :integer
#  channel_id :integer          not null
#  episode_id :integer          not null
#  work_id    :integer          not null
#  started_at :datetime         not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_draft_programs_on_channel_id  (channel_id)
#  index_draft_programs_on_episode_id  (episode_id)
#  index_draft_programs_on_program_id  (program_id)
#  index_draft_programs_on_work_id     (work_id)
#

class DraftProgram < ActiveRecord::Base
  include DraftCommon
  include ProgramCommon

  belongs_to :origin, class_name: "Program", foreign_key: :program_id
end
