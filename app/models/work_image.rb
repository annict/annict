# frozen_string_literal: true
# == Schema Information
#
# Table name: work_images
#
#  id                      :integer          not null, primary key
#  work_id                 :integer          not null
#  user_id                 :integer          not null
#  attachment_file_name    :string           not null
#  attachment_file_size    :integer          not null
#  attachment_content_type :string           not null
#  attachment_updated_at   :datetime         not null
#  aasm_state              :string           default("published"), not null
#  likes_count             :integer          default(0), not null
#  dislikes_count          :integer          default(0), not null
#  created_at              :datetime         not null
#  updated_at              :datetime         not null
#
# Indexes
#
#  index_work_images_on_aasm_state  (aasm_state)
#  index_work_images_on_user_id     (user_id)
#  index_work_images_on_work_id     (work_id)
#

class WorkImage < ApplicationRecord
  include AASM

  has_attached_file :attachment

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  validates :attachment,
    attachment_presence: true,
    attachment_content_type: { content_type: /\Aimage/ }

  belongs_to :work
  belongs_to :user
  has_many :dislikes,
    dependent: :destroy,
    foreign_key: :recipient_id,
    foreign_type: :recipient
  has_many :likes,
    dependent: :destroy,
    foreign_key: :recipient_id,
    foreign_type: :recipient

  scope :sort_by_popular, -> { order("(likes_count - dislikes_count) DESC") }

  before_destroy :change_work_id

  private

  def change_work_id
    return if work.work_image_id != id
    work_image = work.work_images.where.not(id: id).sort_by_popular.first
    work.update_column(:work_image_id, work_image&.id)
  end
end
