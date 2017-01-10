# frozen_string_literal: true
# == Schema Information
#
# Table name: forum_categories
#
#  id                :integer          not null, primary key
#  slug              :string           not null
#  name              :string           not null
#  forum_posts_count :integer          default(0), not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#  name_en           :string           not null
#
# Indexes
#
#  index_forum_categories_on_slug  (slug) UNIQUE
#

class ForumCategory < ApplicationRecord
  has_many :forum_posts, inverse_of: :forum_category, dependent: :destroy
end
