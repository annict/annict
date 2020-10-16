# frozen_string_literal: true
# == Schema Information
#
# Table name: user_programs
#
#  id         :bigint           not null, primary key
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  program_id :bigint           not null
#  user_id    :bigint           not null
#  work_id    :bigint           not null
#
# Indexes
#
#  index_user_programs_on_program_id              (program_id)
#  index_user_programs_on_user_id                 (user_id)
#  index_user_programs_on_user_id_and_program_id  (user_id,program_id) UNIQUE
#  index_user_programs_on_user_id_and_work_id     (user_id,work_id) UNIQUE
#  index_user_programs_on_work_id                 (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (program_id => programs.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => works.id)
#
class UserProgram < ApplicationRecord
  belongs_to :program
  belongs_to :user
  belongs_to :work
end
