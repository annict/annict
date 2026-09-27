# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/channels/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = create(:channel)

    expect {
      delete "/db/channels/#{channel.id}"
    }.not_to change(Channel, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディターではないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel = create(:channel)
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/#{channel.id}"
    }.not_to change(Channel, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディターユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel = create(:channel)
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/#{channel.id}"
    }.not_to change(Channel, :count)

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "管理者ユーザーがログインしているとき、チャンネルを削除できること" do
    user = create(:registered_user, :with_admin_role)
    channel = create(:channel)
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/#{channel.id}"
    }.to change(Channel, :count).by(-1)

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
  end

  it "管理者ユーザーがログインしているとき、存在しないチャンネルを削除しようとするとエラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect {
      delete "/db/channels/non-existent-id"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "管理者ユーザーがログインしているとき、Referrerがある場合はそのページにリダイレクトすること" do
    user = create(:registered_user, :with_admin_role)
    channel = create(:channel)
    login_as(user, scope: :user)

    delete "/db/channels/#{channel.id}", headers: {"HTTP_REFERER" => db_channel_list_path}

    expect(response).to redirect_to(db_channel_list_path)
    expect(flash[:notice]).to eq("削除しました")
  end

  it "管理者ユーザーがログインしているとき、削除されたチャンネルの関連データも削除されること" do
    user = create(:registered_user, :with_admin_role)
    channel = create(:channel)
    program = create(:program, channel:)
    create(:slot, program:, channel:)
    login_as(user, scope: :user)

    delete "/db/channels/#{channel.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Program.where(channel_id: channel.id).count).to eq(0)
    expect(Slot.where(channel_id: channel.id).count).to eq(0)
  end
end
