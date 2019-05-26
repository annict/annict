# frozen_string_literal: true
# == Schema Information
#
# Table name: items
#
#  id                     :bigint(8)        not null, primary key
#  title                  :string           not null
#  detail_page_url        :string           not null
#  asin                   :string           not null
#  ean                    :string
#  amount                 :integer
#  currency_code          :string           default(""), not null
#  offer_amount           :integer
#  offer_currency_code    :string           default(""), not null
#  release_on             :datetime
#  manufacturer           :string           default(""), not null
#  thumbnail_file_name    :string
#  thumbnail_content_type :string
#  thumbnail_file_size    :integer
#  thumbnail_updated_at   :datetime
#  aasm_state             :string           default("published"), not null
#  created_at             :datetime         not null
#  updated_at             :datetime         not null
#  locale                 :string           default("other"), not null
#  image_data             :text
#
# Indexes
#
#  index_items_on_asin    (asin) UNIQUE
#  index_items_on_locale  (locale)
#

class Item < ApplicationRecord
  include ItemImageUploader::Attachment.new(:image)
  include AASM
  include LocaleDetectable
  include ImageUploadable

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  validates :title, presence: true
  validates :detail_page_url, presence: true
  validates :asin, presence: true
end
