# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/staffs/:id", type: :request do
  it "未ログインユーザーは削除できず、ログインページにリダイレクトされること" do
    staff = create(:staff, :not_deleted)

    expect(Staff.count).to eq(1)

    delete "/db/staffs/#{staff.id}"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Staff.count).to eq(1)
  end

  it "一般ユーザーは削除できず、アクセス拒否されること" do
    user = create(:registered_user)
    staff = create(:staff, :not_deleted)

    login_as(user, scope: :user)

    expect(Staff.count).to eq(1)

    delete "/db/staffs/#{staff.id}"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Staff.count).to eq(1)
  end

  it "editor権限ユーザーは削除できず、アクセス拒否されること" do
    user = create(:registered_user, :with_editor_role)
    staff = create(:staff, :not_deleted)

    login_as(user, scope: :user)

    expect(Staff.count).to eq(1)

    delete "/db/staffs/#{staff.id}"
    staff.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Staff.count).to eq(1)
  end

  it "admin権限ユーザーはスタッフをソフトデリートできること" do
    user = create(:registered_user, :with_admin_role)
    staff = create(:staff, :not_deleted)

    login_as(user, scope: :user)

    expect(Staff.count).to eq(1)

    delete "/db/staffs/#{staff.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Staff.count).to eq(0)
  end

  it "存在しないスタッフIDを指定した場合、404エラーになること" do
    user = create(:registered_user, :with_admin_role)

    login_as(user, scope: :user)

    expect do
      delete "/db/staffs/999999"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "既に削除済みのスタッフを指定した場合、404エラーになること" do
    user = create(:registered_user, :with_admin_role)
    staff = create(:staff, :deleted)

    login_as(user, scope: :user)

    expect do
      delete "/db/staffs/#{staff.id}"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
