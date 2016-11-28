# frozen_string_literal: true
# == Schema Information
#
# Table name: character_images
#
#  id                      :integer          not null, primary key
#  character_id            :integer          not null
#  user_id                 :integer          not null
#  attachment_file_name    :string           not null
#  attachment_file_size    :integer          not null
#  attachment_content_type :string           not null
#  attachment_updated_at   :datetime         not null
#  created_at              :datetime         not null
#  updated_at              :datetime         not null
#  aasm_state              :string           default("published"), not null
#  likes_count             :integer          default(0), not null
#  dislikes_count          :integer          default(0), not null
#
# Indexes
#
#  index_character_images_on_character_id  (character_id)
#  index_character_images_on_user_id       (user_id)
#

class CharacterImage < ApplicationRecord
  has_attached_file :attachment

  validates :attachment,
    attachment_presence: true,
    attachment_content_type: { content_type: /\Aimage/ }

  belongs_to :character
  belongs_to :user
end
