# == Schema Information
#
# Table name: db_comments
#
#  id         :integer          not null, primary key
#  body       :text             not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#

class DbComment < ApplicationRecord
end
