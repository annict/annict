# == Schema Information
#
# Table name: edit_requests
#
#  id                    :integer          not null, primary key
#  user_id               :integer          not null
#  kind                  :integer          not null
#  status                :integer          default(1), not null
#  resource_id           :integer
#  resource_type         :string
#  trackable_id          :integer
#  trackable_type        :string
#  draft_resource_params :json             not null
#  title                 :string           not null
#  body                  :text
#  merged_at             :datetime
#  closed_at             :datetime
#  created_at            :datetime         not null
#  updated_at            :datetime         not null
#
# Indexes
#
#  index_edit_requests_on_resource_id_and_resource_type    (resource_id,resource_type)
#  index_edit_requests_on_trackable_id_and_trackable_type  (trackable_id,trackable_type)
#  index_edit_requests_on_user_id                          (user_id)
#

class EditRequest < ActiveRecord::Base
  extend Enumerize

  enumerize :kind, in: { work: 1, episodes: 2, episode: 3, program: 4, item: 5 }
  enumerize :status, in: { opened: 1, merged: 2, closed: 3 }

  belongs_to :user
  belongs_to :resource, polymorphic: true
  belongs_to :trackable, polymorphic: true
  has_many :comments, class_name: "EditRequestComment"
  has_many :images, class_name: "EditRequestImage"

  def closed?
    closed_at.present?
  end
end
