# frozen_string_literal: true

class MakeReviewSimple < ActiveRecord::Migration[5.1]
  def change
    change_column_null :reviews, :title, true
  end
end
