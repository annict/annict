# typed: false
# frozen_string_literal: true

class WorkTagging < ApplicationRecord
  counter_culture :work_tag

  belongs_to :user
  belongs_to :work
  belongs_to :work_tag
end
