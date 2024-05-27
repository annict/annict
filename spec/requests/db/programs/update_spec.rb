# typed: false
# frozen_string_literal: true

describe "PATCH /db/programs/:id", type: :request do
  context "user does not sign in" do
    let!(:channel) { Channel.last }
    let!(:program) { create(:program) }
    let!(:old_program) { program.attributes }
    let!(:program_params) do
      {
        channel_id: channel.id
      }
    end

    it "user can not access this page" do
      patch "/db/programs/#{program.id}", params: {program: program_params}
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(program.channel_id).to eq(old_program["channel_id"])
    end
  end

  context "user who is not editor signs in" do
    let!(:channel) { Channel.last }
    let!(:user) { create(:registered_user) }
    let!(:program) { create(:program) }
    let!(:old_program) { program.attributes }
    let!(:program_params) do
      {
        channel_id: channel.id
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      patch "/db/programs/#{program.id}", params: {program: program_params}
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(program.channel_id).to eq(old_program["channel_id"])
    end
  end

  context "user who is editor signs in" do
    let!(:channel) { Channel.first }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:program) { create(:program) }
    let!(:old_program) { program.attributes }
    let!(:program_params) do
      {
        channel_id: channel.id
      }
    end
    let!(:attr_names) { program_params.keys }

    before do
      login_as(user, scope: :user)
    end

    it "user can update program" do
      expect(program.channel_id).to eq(old_program["channel_id"])

      patch "/db/programs/#{program.id}", params: {program: program_params}
      program.reload

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("更新しました")

      expect(program.channel_id).to eq(channel.id)
    end
  end
end
