# == Schema Information
#
# Table name: edit_request_comments
#
#  id              :integer          not null, primary key
#  edit_request_id :integer          not null
#  user_id         :integer          not null
#  body            :text             not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_edit_request_comments_on_edit_request_id  (edit_request_id)
#  index_edit_request_comments_on_user_id          (user_id)
#

class EditRequestComment < ActiveRecord::Base
  belongs_to :edit_request
  belongs_to :user

  validates :body, presence: true

  after_create :create_participant

  private

  def create_participant
    edit_request.participants.where(user: user).first_or_create
  end
end
