# frozen_string_literal: true
# == Schema Information
#
# Table name: anime_images
#
#  id                      :bigint           not null, primary key
#  asin                    :string           default(""), not null
#  attachment_content_type :string
#  attachment_file_name    :string
#  attachment_file_size    :integer
#  attachment_updated_at   :datetime
#  color_rgb               :string           default("255,255,255"), not null
#  copyright               :string           default(""), not null
#  image_data              :text             not null
#  created_at              :datetime         not null
#  updated_at              :datetime         not null
#  user_id                 :bigint           not null
#  work_id                 :bigint           not null
#
# Indexes
#
#  index_anime_images_on_user_id  (user_id)
#  index_anime_images_on_work_id  (work_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_id => animes.id)
#

class AnimeImage < ApplicationRecord
  include WorkImageUploader::Attachment.new(:image)
  include ImageUploadable

  validates :copyright, presence: true

  belongs_to :work, touch: true
  belongs_to :user

  # Disable until the Paperclip -> Shrine migration is done
  # after_save :set_color_rgb

  def colors
    return @colors if @colors.present?

    colors = Miro::DominantColors.new(image_url(:master))
    @colors = colors.to_rgb.map { |c| c.join(",") }
    @colors
  end

  def text_color_rgb(light: "255,255,255", dark: "0,0,0")
    # https://stackoverflow.com/questions/3942878/how-to-decide-font-color-in-white-or-black-depending-on-background-color
    red, green, blue = color_rgb.split(",").map(&:to_i)
    (red * 0.299 + green * 0.587 + blue * 0.114) > 186 ? dark : light
  end

  def color_hex
    color_rgb.split(",").map { |i| i.to_i.to_s(16).rjust(2, "0") }.join
  end

  private

  def set_color_rgb
    colors = Miro::DominantColors.new(image_url(:master))
    color_rgb = colors.to_rgb.map { |c| c.join(",") }.first
    update_column(:color_rgb, color_rgb)
  end
end
