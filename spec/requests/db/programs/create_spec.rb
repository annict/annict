# typed: false
# frozen_string_literal: true

describe "POST /db/works/:work_id/programs", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.first }
    let!(:work) { create(:work) }
    let!(:form_params) do
      {
        rows: "#{channel.id},2020-04-01 0:00"
      }
    end

    it "user can not access this page" do
      post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Program.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:channel) { Channel.first }
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user) }
    let!(:form_params) do
      {
        rows: "#{channel.id},2020-04-01 0:00"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Program.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:channel) { Channel.first }
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:form_params) do
      {
        rows: "#{channel.id},2020-04-01 0:00"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create program" do
      expect(Program.all.size).to eq(0)

      post "/db/works/#{work.id}/programs", params: {deprecated_db_program_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Program.all.size).to eq(1)
      program = Program.last

      expect(program.channel_id).to eq(channel.id)
      expect(program.started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
    end
  end
end
