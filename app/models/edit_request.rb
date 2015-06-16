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
#  published_at        :datetime
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

  attr_accessor :publisher

  belongs_to :user
  belongs_to :draft_resource, polymorphic: true
  has_many :comments, class_name: "EditRequestComment"
  has_many :participants, class_name: "EditRequestParticipant"

  validates :title, presence: true

  after_create :create_participant

  aasm do
    state :opened, initial: true
    state :published
    state :closed

    event :publish do
      transitions from: :opened, to: :published do
        after do
          publish_edit_request!
          participants.where(user: publisher).first_or_create
        end
      end
    end

    event :close do
      transitions from: [:opened, :published], to: :closed
    end
  end

  def kind
    draft_resource.class.name.underscore
  end

  private

  def publish_edit_request!
    attrs = draft_resource.slice(*draft_resource.class::PUBLISH_FIELDS)

    if draft_resource.origin.present?
      draft_resource.origin.update(attrs)
    else
      origin_class = draft_resource.class.reflections["origin"]
                                   .class_name.constantize
      origin_class.create(attrs)
    end

    update_column(:published_at, Time.now)
  end

  def create_participant
    participants.create do |p|
      p.user = user
    end
  end
end
