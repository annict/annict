# == Schema Information
#
# Table name: edit_request_participants
#
#  id              :integer          not null, primary key
#  edit_request_id :integer          not null
#  user_id         :integer          not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_edit_request_participants_on_edit_request_id              (edit_request_id)
#  index_edit_request_participants_on_edit_request_id_and_user_id  (edit_request_id,user_id) UNIQUE
#  index_edit_request_participants_on_user_id                      (user_id)
#

class EditRequestParticipant < ActiveRecord::Base
  belongs_to :edit_request
  belongs_to :user
end
