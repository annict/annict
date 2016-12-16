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

  attr_accessor :proposer

  belongs_to :user
  belongs_to :draft_resource, polymorphic: true
  has_many :comments, class_name: "EditRequestComment", dependent: :destroy
  has_many :participants, class_name: "EditRequestParticipant", dependent: :destroy

  delegate :origin, to: :draft_resource

  validates :title, presence: true

  after_create :create_participant
  after_create :create_db_activity
  after_create :notify_new_edit_request

  aasm do
    state :opened, initial: true
    state :published
    state :closed

    event :publish do
      transitions from: :opened, to: :published do
        after do
          publish_edit_request!
          participants.where(user: proposer).first_or_create

          DbActivity.create do |a|
            a.user = proposer
            a.trackable = self
            a.action = "edit_requests.publish"
          end
        end
      end
    end

    event :close do
      transitions from: :opened, to: :closed do
        after do
          participants.where(user: proposer).first_or_create
          update_column(:closed_at, Time.now)

          DbActivity.create do |a|
            a.user = proposer
            a.trackable = self
            a.action = "edit_requests.close"
          end
        end
      end
    end
  end

  def kind
    draft_resource.class.name.underscore.to_sym
  end

  def draft_resource_parent
    case draft_resource.class.name
    when "DraftEpisode", "DraftItem", "DraftMultipleEpisode", "DraftProgram"
      draft_resource.work
    when "DraftCast", "DraftStaff"
      draft_resource.person
    end
  end

  def db_activities
    condition = <<-SQL
      (recipient_type = 'EditRequest' AND recipient_id = ?) OR
      (trackable_type = 'EditRequest' AND trackable_id = ?)
    SQL
    DbActivity.where(condition, id, id)
  end

  private

  def publish_edit_request!
    if draft_resource.instance_of?(DraftMultipleEpisode)
      work = draft_resource.work
      hash = draft_resource.to_episode_hash
      Episode.create_from_multiple_episodes(work, hash)
    else
      attrs = draft_resource.slice(*draft_resource.class::PUBLISH_FIELDS)

      if draft_resource.origin.present?
        draft_resource.origin.update(attrs)
      else
        origin_class = draft_resource.class.reflections["origin"].
                       class_name.constantize
        origin_class.create!(attrs)
      end
    end

    update_column(:published_at, Time.now)
  end

  def create_participant
    participants.create do |p|
      p.user = user
    end
  end

  def create_db_activity
    DbActivity.create do |a|
      a.user = user
      a.trackable = self
      a.action = "edit_requests.create"
    end
  end

  def notify_new_edit_request
    return unless Rails.env.production?

    webhook_url = ENV.fetch("ANNICT_SLACK_WEBHOOK_URL_FOR_NOTIFICATIONS")
    options = { channel: "#notifications", username: "Notifier", icon_emoji: ":annict:" }
    host = ENV.fetch("ANNICT_URL")
    url = Rails.application.routes.url_helpers.db_edit_request_url(self, host: host)
    message = "<!channel> #{user.profile.name}さんが編集リクエストを作成しました。 #{url}"

    notifier = Slack::Notifier.new(webhook_url, options)
    notifier.ping(message)
  end
end
