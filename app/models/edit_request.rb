# == Schema Information
#
# Table name: edit_requests
#
#  id                  :integer          not null, primary key
#  user_id             :integer          not null
#  draft_resource_id   :integer          not null
#  draft_resource_type :string           not null
#  title               :string           not null
#  body                :text
#  aasm_state          :string           default("opened"), not null
#  merged_at           :datetime
#  closed_at           :datetime
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#
# Indexes
#
#  index_edit_requests_on_user_id  (user_id)
#  index_er_on_drid_and_drtype     (draft_resource_id,draft_resource_type)
#

class EditRequest < ActiveRecord::Base
  include AASM

  belongs_to :user
  belongs_to :draft_resource, polymorphic: true
  has_many :comments, class_name: "EditRequestComment"

  aasm do
    state :opened, initial: true
    state :merged
    state :closed

    event :merge do
      transitions from: :opened, to: :merged
    end

    event :close do
      transitions from: [:opened, :merged], to: :closed
    end
  end

  def kind
    draft_resource.class.name.underscore
  end
end
