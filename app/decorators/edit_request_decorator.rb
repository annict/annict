class EditRequestDecorator < Draper::Decorator
  delegate_all

  def edit_path
    case object.kind
    when "work"
      h.edit_db_works_edit_request_path(object)
    when "episodes"
      h.edit_db_work_episodes_edit_request_path(object.trackable, object)
    when "episode"
      h.edit_db_work_episode_edit_request_path(object.trackable, object.resource, object)
    when "program"
      if object.resource.present?
        args = [object.trackable, object.resource, object]
        h.edit_db_work_program_edit_request_path(*args)
      else
        h.edit_db_work_programs_edit_request_path(object.trackable, object)
      end
    end
  end

  def update_path
    case object.kind
    when "work"
      h.db_works_edit_request_path(object)
    end
  end

  def status_label
    if status.opened?
      h.content_tag :span, "オープン", class: "label label-success"
    elsif status.merged?
      h.content_tag :span, "反映済み", class: "label label-info"
    elsif status.closed?
      h.content_tag :span, "クローズ", class: "label label-danger"
    end
  end

  def to_diffable_draft_resource
    case object.kind
    when "work"
      hash["media"] = Work.media.find_value(hash["media"]).text
    when "program"
      to_diffable_program!
    end
  end

  private

  def to_diffable_program!
    hash = {}

    object.draft_resource_params.each do |key, val|
      hash[key] = case key
                  when "channel_id"
                    { data: val, value: Channel.find(val).name }
                  when "episode_id"
                    episode = Episode.find(val)
                    episode_path = h.work_episode_path(episode.work, episode)
                    episode_title = episode.decorate.title_with_number

                    {
                      data: val,
                      value: h.link_to(episode_title, episode_path, target: "_blank")
                    }
                  else
                    { data: val, value: val }
                  end
    end

    hash
  end
end
