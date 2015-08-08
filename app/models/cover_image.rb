# == Schema Information
#
# Table name: cover_images
#
#  id         :integer          not null, primary key
#  work_id    :integer          not null
#  file_name  :string           not null
#  location   :string           not null
#  created_at :datetime
#  updated_at :datetime
#

class CoverImage < ActiveRecord::Base
  belongs_to :work
end
