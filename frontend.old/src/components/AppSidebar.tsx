import { NavLink } from "@/components/NavLink";
import { FileText, Plus, User, Zap } from "lucide-react";

const AppSidebar = () => {
  const menuItems = [
    {
      title: "Novo Fluxo",
      icon: Plus,
      href: "/app/novo",
    },
    {
      title: "Gerenciar Matérias",
      icon: FileText,
      href: "/app/materias",
    },
    {
      title: "Minha Conta",
      icon: User,
      href: "/app/conta",
    },
  ];

  return (
    <aside className="w-[280px] m-4 h-[calc(100vh-32px)] bg-card rounded-xl shadow-card border border-border/50 flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-border/50">
        <div className="flex items-center gap-2">
          <Zap className="h-7 w-7 text-primary" />
          <span className="text-2xl font-bold text-primary-dark">organiQ</span>
        </div>
      </div>

      {/* Menu Items */}
      <nav className="flex-1 p-4">
        <ul className="space-y-2">
          {menuItems.map((item) => {
            const Icon = item.icon;
            return (
              <li key={item.href}>
                <NavLink
                  to={item.href}
                  end
                  className="flex items-center gap-3 px-4 py-3 rounded-lg text-foreground/70 hover:bg-accent/5 hover:text-foreground"
                  activeClassName="bg-primary/10 text-primary font-medium border-l-4 border-primary pl-[14px]"
                >
                  <Icon className="h-5 w-5" />
                  <span>{item.title}</span>
                </NavLink>
              </li>
            );
          })}
        </ul>
      </nav>
    </aside>
  );
};

export default AppSidebar;
