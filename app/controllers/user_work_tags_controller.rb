# frozen_string_literal: true

class UserWorkTagsController < ApplicationController
  def show(username, id)
    @user = User.published.find_by!(username: username)
    @tag = @user.work_tags.find_by!(name: id)
    @taggable = @user.work_taggables.find_by(work_tag: @tag)
    @works = Work.joins(:work_taggings).merge(WorkTagging.where(user: @user, work_tag: @tag))

    store_page_params(works: @works)
  end
end
