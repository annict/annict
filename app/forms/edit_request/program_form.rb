class EditRequest::ProgramForm
  include Virtus.model
  include ActiveModel::Model

  attr_reader :edit_request_id, :program, :user, :work

  attribute :program_channel_id, Integer
  attribute :program_episode_id, Integer
  attribute :program_started_at, String
  attribute :edit_request_title, String
  attribute :edit_request_body, String

  def work=(work)
    @work ||= work
  end

  def program=(program)
    @program ||= program
  end

  def user=(user)
    @user ||= user
  end

  def edit_request_id=(id)
    @edit_request_id ||= id
  end

  def new_attributes=(program)
    self.program_channel_id = program.channel_id
    self.program_episode_id = program.episode_id
    self.program_started_at = program.started_at.to_time.strftime("%Y/%m/%d %H:%M")
  end

  def edit_attributes=(edit_request)
    self.edit_request_id = edit_request.id
    self.program_channel_id = edit_request.draft_resource_params["channel_id"]
    self.program_episode_id = edit_request.draft_resource_params["episode_id"]
    self.program_started_at = edit_request.draft_resource_params["started_at"]
    self.edit_request_title = edit_request.title
    self.edit_request_body = edit_request.body
  end

  def create_path(controller)
    if program.present?
      controller.db_work_program_edit_requests_path(work, program)
    else
      controller.db_work_programs_edit_requests_path(work)
    end
  end

  def update_path(controller)
    if program.present?
      controller.db_work_program_edit_request_path(work, program, edit_request_id)
    else
      controller.db_work_programs_edit_request_path(work, edit_request_id)
    end
  end

  def persisted?
    edit_request_id.present?
  end

  def valid?
    program = Program.new do |p|
      p.channel_id = program_channel_id
      p.episode_id = program_episode_id
      p.work_id = work.id
      p.started_at = program_started_at
    end

    program.valid?

    {}.merge(program.errors).each do |key, errors|
      self.errors.add("program_#{key}", *errors)
    end

    begin
      DateTime.parse(program_started_at)
    rescue
      self.errors.add(:program_started_at, "のフォーマットが正しくありません。")
    end

    self.errors.blank?
  end

  def save
    return false unless valid?

    draft_resource_params = program_attrs.inject({}) do |hash, attr|
      hash.merge(attr => send("program_#{attr}"))
    end

    edit_request = if persisted?
      EditRequest.find(edit_request_id)
    else
      EditRequest.new
    end

    edit_request.attributes = {
      user: user,
      trackable: work,
      kind: :program,
      resource: program,
      draft_resource_params: draft_resource_params,
      title: edit_request_title.presence || "無題",
      body: edit_request_body
    }

    edit_request.save(validate: false)
    self.edit_request_id = edit_request.id

    true
  end

  def persisted?
    edit_request_id.present?
  end

  private

  def program_attrs
    attributes.keys.select { |attr| /\Aprogram_/ === attr }.map do |attr|
      attr.to_s.sub(/\Aprogram_/, "")
    end
  end
end
