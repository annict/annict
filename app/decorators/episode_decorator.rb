class EpisodeDecorator < ApplicationDecorator
  include EpisodeDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_episode_path(work, self)
    else
      h.new_db_work_draft_episode_path(work, id: id)
    end

    h.link_to name, path, options
  end
end
