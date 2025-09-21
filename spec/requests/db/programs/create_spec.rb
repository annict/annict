# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/programs", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = Channel.first
    work = create(:work)
    form_params = {
      rows: "#{channel.id},2020-04-01 0:00"
    }

    post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Program.all.size).to eq(0)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    channel = Channel.first
    work = create(:work)
    user = create(:registered_user)
    form_params = {
      rows: "#{channel.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Program.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしているとき、放送予定を作成できること" do
    channel = Channel.first
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{channel.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    expect(Program.all.size).to eq(0)

    post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Program.all.size).to eq(1)

    program = Program.last
    expect(program.channel_id).to eq(channel.id)
    expect(program.started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
  end

  it "管理者権限を持つユーザーがログインしているとき、放送予定を作成できること" do
    channel = Channel.first
    work = create(:work)
    user = create(:registered_user, :with_admin_role)
    form_params = {
      rows: "#{channel.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    expect(Program.all.size).to eq(0)

    post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Program.all.size).to eq(1)

    program = Program.last
    expect(program.channel_id).to eq(channel.id)
    expect(program.started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
  end

  it "複数の放送予定を一度に作成できること" do
    channel1 = Channel.first
    channel2 = Channel.second
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{channel1.id},2020-04-01 0:00\n#{channel2.id},2020-04-02 1:00"
    }

    login_as(user, scope: :user)

    expect(Program.all.size).to eq(0)

    post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Program.all.size).to eq(2)

    programs = Program.order(:started_at)
    expect(programs[0].channel_id).to eq(channel1.id)
    expect(programs[0].started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
    expect(programs[1].channel_id).to eq(channel2.id)
    expect(programs[1].started_at.to_s).to eq(Time.zone.parse("2020-04-01 16:00").to_s)
  end

  it "不正な時刻形式が指定されたとき、バリデーションエラーになること" do
    channel = Channel.first
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{channel.id},invalid-time"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(422)
    expect(Program.all.size).to eq(0)
  end

  it "存在しない作品IDが指定されたとき、404エラーになること" do
    channel = Channel.first
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{channel.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    post "/db/works/999999/programs", params: {deprecated_db_program_rows_form: form_params}

    expect(response.status).to eq(404)
  end
end
